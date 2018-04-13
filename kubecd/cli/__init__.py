import os
import subprocess
import sys
from typing import List

import click
import ruamel.yaml
import ruamel.yaml.error

from kubecd.updates import find_updates_for_env
from .. import environments
from .. import __version__


def print_version(ctx, param, value):
    if not value or ctx.resilient_parsing:
        return
    click.echo('kubecd version ' + __version__)
    ctx.exit()


@click.group()
@click.option('--environments-file', '-f',
              envvar='KCD_ENVIRONMENTS',
              type=click.Path(),
              help='Your environments.yaml file',
              default='environments.yaml')
@click.option('--version', is_flag=True, callback=print_version, expose_value=False, is_eager=True,
              help='Show version and exit.')
@click.pass_context
def main(ctx, environments_file):
    ctx.obj = {
        'environments_file': environments_file
    }


@main.group(help='commands for managing environments')
def env():
    pass


def load_envs(file_name: str) -> List[environments.Environment]:
    try:
        environments.load(file_name)
        return environments.as_list()
    except ruamel.yaml.error.YAMLError as e:
        click.echo(click.style('ERROR:', fg='red') + ' could not read "{}": {}'.format(file_name, str(e)))
        sys.exit(1)


def one_or_more_envs(env_name: str, file_name: str) -> List[environments.Environment]:
    load_envs(file_name)
    if env_name is not None:
        try:
            return [environments.get_environment(env_name)]
        except KeyError:
            click.echo(click.style('ERROR:', fg='red') + ' no such environment: {}'.format(env_name))
            sys.exit(1)
    return environments.as_list()


def one_env(env_name: str, file_name: str):
    load_envs(file_name)
    try:
        return environments.get_environment(env_name)
    except KeyError:
        click.echo(click.style('ERROR:', fg='red') + ' no such environment: {}'.format(env_name))
        sys.exit(1)


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


@env.command(name='apply', help='apply current state for an environment')
@click.argument('ENV_NAME')
@click.option('--dry-run', '-n', is_flag=True, default=False, help='dry run mode - only print commands')
@click.pass_context
def apply_env(ctx, env_name, dry_run):
    environment = one_env(env_name, file_name=ctx.obj['environments_file'])
    commands_to_run = []
    commands_to_run.extend(environment.init_commands(dry_run=dry_run))
    commands_to_run.extend(environment.deploy_commands(dry_run=dry_run))
    for cmd in commands_to_run:
        click.secho(' '.join(cmd), fg='yellow')
        cmd_status = subprocess.call(cmd)
        if cmd_status != 0:
            click.secho('Command exited with non-0 status {}'.format(cmd_status), fg='red', bold=True)
            break


@env.command(name='list', help='list all environment names')
@click.pass_context
def list_envs(ctx):
    for environment in environments.load(ctx.obj['environments_file']):
        click.echo(environment.name)


@env.command(name='dump', help='dump one or all environments')
@click.argument('ENV_NAME', required=False)
@click.pass_context
def dump_envs(ctx, env_name=None):
    for environment in one_or_more_envs(env_name, ctx.obj['environments_file']):
        click.echo(click.style('Environment: ', fg='green') + click.style(environment.name, fg='green', bold=True))
        for cmd in environment.init_commands(dry_run=False):
            click.secho(' '.join(cmd), fg='yellow')
        for cmd in environment.deploy_commands(dry_run=False):
            click.echo(' '.join(cmd))
        click.echo('')


@main.command(name='check-updates', help='check for image updates')
@click.argument('ENV_NAME', required=False)
@click.option('--patch', is_flag=True, help='patch release files with new tags')
@click.pass_context
def check_updates(ctx, env_name=None, patch=False):
    for environment in one_or_more_envs(env_name, ctx.obj['environments_file']):
        updates = find_updates_for_env(environment)
        for release_file in updates:
            mod_yaml = None
            if patch:
                mod_yaml = load_yaml(release_file)
            for update in updates[release_file]:
                click.echo('\t{file}: release "{release}" image "{image}" tag "{tag_value}" {tag_from} -> {tag_to}'.format(
                    file=release_file,
                    release=update['release'],
                    image=update['image_repo'],
                    tag_value=update['tag_value'],
                    tag_from=update['old_tag'],
                    tag_to=update['new_tag'],
                ))
                if patch:
                    for yr in mod_yaml['releases']:
                        if yr['name'] == update['release']:
                            if 'values' not in yr:
                                yr['values'] = []
                            found_val = False
                            for yv in yr['values']:
                                if yv['key'] == update['tag_value']:
                                    yv['value'] = update['new_tag']
                                    found_val = True
                                    break
                            if not found_val:
                                yr['values'].append({'key': update['tag_value'], 'value': update['new_tag']})
            if patch:
                save_yaml(mod_yaml, release_file)


@main.command(name='indent', help='canonically indent a YAML file')
@click.argument('FILENAME')
def indent_file(filename):
    try:
        save_yaml(load_yaml(filename), filename)
    except ruamel.yaml.error.YAMLError as e:
        click.echo(click.style('ERROR:', fg='red') + ' invalid YAML file "{}": {}'.format(filename, str(e)))


if __name__ == '__main__':
    main()
