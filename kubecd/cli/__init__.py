import argparse
import os
import subprocess
import sys
from typing import List

import argcomplete
import logging
import ruamel.yaml
import ruamel.yaml.error
from argcomplete import FilesCompleter
from blessings import Terminal
from ruamel.yaml import YAMLError

from kubecd import __version__, helm, updates, github
from kubecd import model
from kubecd.updates import find_updates_for_env, find_updates_for_release

t = Terminal()
logger = logging.getLogger(__name__)


class CliError(Exception):
    pass


def parser(prog='kcd') -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog=prog)
    p.add_argument(
        '-f', '--config-file',
        help='KubeCD config file file (default $KUBECD_CONFIG or "environments.yaml")',
        metavar='FILE',
        default=os.getenv('KUBECD_CONFIG', 'environments.yaml')).completer = FilesCompleter(
            allowednames=('.yaml', '.yml'), directories=False)
    p.add_argument(
        '--version',
        help='Show version and exit.',
        action='version',
        version='kubecd ' + __version__)
    p.add_argument('--verbose', '-v', help='Increase verbosity level', action='count', default=0)

    s = p.add_subparsers(dest='command', title='Subcommands', description='Use one of these sub-commands:')

    apply = s.add_parser('apply', help='apply changes to Kubernetes')
    apply.add_argument('--dry-run', '-n', help='dry run mode, only print commands', action='store_true', default=False)
    apply.add_argument('--debug', help='run helm with --debug', action='store_true', default=False)
    apply.add_argument('--release', '-r', help='apply only this release')
    apply.add_argument('--all-environments', '-a', help='apply all environments', action='store_true', default=False)
    apply.add_argument('env', nargs='?',
                       help='name of environment to apply, must be specified unless --all-environments is')
    apply.set_defaults(func=apply_env)

    diff = s.add_parser('diff', help='show diffs between running and git release')
    diff.add_argument('--release', '-r', help='which release to diff')
    diff.add_argument('env', nargs='?', help='name of environment')
    diff.set_defaults(func=diff_release)

    poll_p = s.add_parser('poll', help='poll for new images in registries')
    poll_p.add_argument('--patch', '-p', help='patch releases.yaml files with updated version', action='store_true')
    poll_p.add_argument('--release', '-r', help='poll this specific release')
    poll_p.add_argument('--image', '-i', help='poll releases using this image')
    poll_p.add_argument('env', nargs='?', help='name of environment to poll')
    poll_p.set_defaults(func=poll_registries)

    dump_p = s.add_parser('dump', help='dump commands for one or all environments')
    dump_p.add_argument('--all-environments', '-a', help='dump all environments', action='store_true', default=False)
    dump_p.add_argument('env', nargs='?', help='name of environment to dump')
    dump_p.set_defaults(func=dump_env)

    list_p = s.add_parser('list', help='list environments or releases')
    list_p.add_argument('kind', help='what to list', choices=['env', 'release'])
    list_p.set_defaults(func=list_kind)

    indent_p = s.add_parser('indent', help='canonically indent YAML files')
    indent_p.add_argument('files', nargs='+', help='file[s] to indent')
    indent_p.set_defaults(func=indent_file)

    observe = s.add_parser('observe', help='observe a new image version')
    observe.add_argument('--image', '-i', help='the image, including tag', metavar='IMAGE:TAG')
    observe.add_argument('--patch', help='patch release files with updated tags',
                         action='store_true', default=False)
    observe.add_argument('--submit-pr', help='submit a pull request with the updated tags',
                         action='store_true', default=False)
    observe.set_defaults(func=observe_new_image)

    completion_p = s.add_parser('completion', help='print shell completion script')
    completion_p.set_defaults(func=print_completion, prog=prog)
    return p


def diff_release(config_file: str, release: str, env: str, **kwargs):
    kcd_config = load_model(config_file)
    e = kcd_config.env_by_name(env)
    r = e.named_release(release)
    cmd = helm.generate_helm_diff_argv(r, e, r.from_file)
    s = subprocess.call(cmd)
    if s != 0:
        raise CliError('Command "{cmd}" exited with status {status}'.format(cmd=' '.join(cmd), status=s))


def observe_new_image(config_file: str, image: str, patch: bool, submit_pr: bool, **kwargs):
    kcd_config = load_model(config_file)
    image_repo, image_tag = image.split(':')
    image_index = kcd_config.image_index
    patched_files = []
    if image_repo in image_index:
        for release in image_index[image_repo]:
            update_list = updates.release_wants_tag_update(release, image_tag)
            for update in update_list:
                print('release: {release} tagValue: {tag}'.format(release=update.release.name, tag=update.tag_value))
            if len(update_list) > 0 and patch:
                patch_releases_file(release.from_file, update_list)
                patched_files.append(release.from_file)
    if submit_pr and len(patched_files) > 0:
        github.pr_from_files(patched_files, 'Observed image {image}'.format(image=image))


def print_completion(prog, **kwargs):
    shell = os.path.basename(os.getenv('SHELL'))
    if shell == 'bash' or shell == 'tcsh':
        sys.stdout.write(argcomplete.shellcode(prog, shell=shell))


def apply_env(config_file, dry_run, debug, all_environments=False, env=None, release=None, **kwargs):
    target_envs = one_or_all_envs(env, all_environments, file_name=config_file)
    commands_to_run = model.config().init_commands()
    for environment in target_envs:
        logger.info('Collecting commands for environment "%s"', environment.name)
        init_cmds = environment.init_commands(dry_run=dry_run)
        deploy_cmds = helm.deploy_commands(environment, dry_run=dry_run, limit_to_release=release, debug=debug)
        if len(deploy_cmds) > 0:
            commands_to_run.extend(init_cmds)
            commands_to_run.extend(deploy_cmds)
    for cmd in commands_to_run:
        print('{t.yellow}{cmd}{t.normal}'.format(cmd=' '.join(cmd), t=t))
        logger.debug('Executing: "%s"', ' '.join(cmd))
        cmd_status = subprocess.call(cmd)
        if cmd_status != 0:
            raise CliError('Command "{cmd}" exited with non-0 status {status}'.format(cmd=cmd, status=cmd_status))


def dump_env(config_file, all_environments=False, env=None, **kwargs):
    target_envs = one_or_all_envs(env, all_environments, file_name=config_file)
    for cmd in model.config().init_commands():
        print('{t.yellow}{cmd}{t.normal}'.format(cmd=' '.join(cmd), t=t))
    for environment in target_envs:
        print('{t.green}Environment:{t.normal} {env_name}'.format(env_name=environment.name, t=t))
        for cmd in environment.init_commands(dry_run=False):
            print('{t.yellow}{cmd}{t.normal}'.format(cmd=' '.join(cmd), t=t))
        for cmd in helm.deploy_commands(environment, dry_run=False):
            print(' '.join(cmd))
        print('')


def list_kind(config_file, kind, **kwargs):
    if kind == 'env' or kind == 'environment':
        for environment in load_model(config_file):
            print(environment.name)
    elif kind == 'release':
        for environment in load_model(config_file):
            for release in environment.all_releases:
                print('{env}/{release}'.format(env=environment.name, release=release.name))
    else:
        raise CliError('unknown kind "{}"'.format(kind))


def poll_registries(config_file, all_environments=False, env=None, release=None, patch=False, **kwargs):
    target_envs = one_or_all_envs(env, all_environments, file_name=config_file)
    for environment in target_envs:
        if release is None:
            logger.info('polling environment: "%s"', environment.name)
            file_updates = find_updates_for_env(environment)
        else:
            release_obj = environment.named_release(release)
            if release_obj is None:
                raise CliError('no release called "{} in environment "{}""'.format(release, env))
            logger.info('polling release: "%s/%s"', environment.name, release_obj.name)
            file_updates = find_updates_for_release(release_obj, environment)
        for release_file, update_list in file_updates.items():
            for update in update_list:
                print(
                    '{env}/{release}:\n\tfile: {file}\n\timage: {image}\n\ttag: {tag_from} -> {tag_to}'.format(
                        env=environment.name,
                        file=release_file,
                        release=update.release.name,
                        image=update.image_repo,
                        tag_from=update.old_tag,
                        tag_to=update.new_tag,
                    ))
            if patch:
                patch_releases_file(release_file, update_list)


def patch_releases_file(releases_file: str, updates_list: List[updates.ImageUpdate]):
    logger.debug('loading releases file: "%s"', releases_file)
    mod_yaml = load_yaml(releases_file)
    for update in updates_list:
        for mod_rel in mod_yaml['releases']:
            if mod_rel['name'] == update.release.name:
                if 'values' not in mod_rel:
                    mod_rel['values'] = []
                found_val = False
                for yv in mod_rel['values']:
                    if yv['key'] == update.tag_value:
                        logger.debug('patching value "{}"'.format(update.tag_value))
                        yv['value'] = update.new_tag
                        found_val = True
                        break
                if not found_val:
                    mod_rel['values'].append({'key': update.tag_value, 'value': update.new_tag})
    logger.debug('saving patched file: {file}'.format(file=releases_file))
    save_yaml(mod_yaml, releases_file)


def indent_file(files, **kwargs):
    for filename in files:
        try:
            save_yaml(load_yaml(filename), filename)
        except YAMLError as e:
            raise CliError('invalid YAML file "{}": {}'.format(filename, str(e)))


def load_envs(file_name: str) -> List[model.Environment]:
    try:
        load_model(file_name)
        return model.as_list()
    except ruamel.yaml.error.YAMLError as e:
        raise CliError('could not read "{}": {}'.format(file_name, str(e)))


def one_or_all_envs(env_name: str, all_envs: bool, file_name: str) -> List[model.Environment]:
    load_envs(file_name)
    if env_name is None and all_envs is False:
        raise CliError('need to specify either an environment name or --all')
    if env_name is not None:
        try:
            return [model.get_environment(env_name)]
        except KeyError:
            raise CliError('no such environment: {}'.format(env_name))
    return model.as_list()


def load_model(file_name: str):
    try:
        return model.load(file_name)
    except FileNotFoundError as e:
        raise CliError('file not found: {}'.format(e.filename))


def load_yaml(file_name: str):
    yaml = ruamel.yaml.YAML(typ='rt')
    yaml.indent(sequence=4, offset=2)
    yaml.preserve_quotes = True
    with open(file_name, 'r') as rf:
        return yaml.load(rf)


def save_yaml(mod_yaml, file_name: str):
    yaml = ruamel.yaml.YAML(typ='rt')
    yaml.indent(sequence=4, offset=2)
    yaml.preserve_quotes = True
    tmp_file = file_name + '.new'
    with open(tmp_file, 'w') as wf:
        yaml.dump(mod_yaml, wf)
    os.rename(tmp_file, file_name)


def verbose_log_level(v):
    if v == 0:
        return logging.WARNING
    if v == 1:
        return logging.INFO
    return logging.DEBUG


def main():
    p = parser()
    argcomplete.autocomplete(p)
    args = p.parse_args()
    kwargs = args.__dict__
    if 'func' not in kwargs:
        p.print_help(sys.stderr)
        sys.exit(1)
    func = kwargs['func']
    del (kwargs['func'])
    logging.basicConfig(stream=sys.stderr,
                        format='{levelname} {message}',
                        style='{',
                        level=verbose_log_level(args.verbose))
    try:
        func(**kwargs)
    except CliError as e:
        print('{t.red}ERROR{t.normal}: {msg}'.format(msg=str(e), t=t), file=sys.stderr)


if __name__ == '__main__':
    main()
