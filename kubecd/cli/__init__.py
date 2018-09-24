import argparse
import os
import sys

import argcomplete
import logging
from argcomplete import FilesCompleter
from blessings import Terminal

from .. import __version__
from .commands import (
    apply_env,
    dump_env,
    indent_file,
    init_contexts,
    json2yaml,
    lint_environment,
    list_kind,
    observe_new_image,
    poll_registries,
    use_env_context,
    CliError,
)

t = Terminal()
logger = logging.getLogger(__name__)


def parser(prog='kcd') -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog=prog)
    # not used yet, but left here as a reminder to not steal the -c flag
    # p.add_argument('-c', '--config-file', help='path to configuration file')
    p.add_argument(
        '-f', '--environments-file',
        help='KubeCD config file file (default $KUBECD_ENVIRONMENTS or "environments.yaml")',
        metavar='FILE',
        default=os.getenv('KUBECD_ENVIRONMENTS', 'environments.yaml')).completer = FilesCompleter(
            allowednames=('.yaml', '.yml'), directories=False)
    p.add_argument(
        '--version',
        help='Show version and exit.',
        action='version',
        version='kubecd ' + __version__)
    p.add_argument('--verbose', '-v', help='Increase verbosity level', action='count', default=0)

    s = p.add_subparsers(dest='command', title='Subcommands', description='Use one of these sub-commands:')

    apply = s.add_parser('apply',
                         help='apply changes to Kubernetes')
    apply.add_argument('--dry-run', '-n', action='store_true', default=False,
                       help='dry run mode, only print commands')
    apply.add_argument('--debug', action='store_true', default=False,
                       help='run helm with --debug')
    apply.add_argument('--releases', '-r', action='append',
                       help='apply only these releases')
    apply.add_argument('--cluster', '-c', nargs='?', metavar='CLUSTER',
                       help='apply all environments in CLUSTER')
    apply.add_argument('--init', action='store_true', default=False,
                       help='Initialize credentials and contexts')
    apply.add_argument('env_name', nargs='?', metavar='ENV',
                       help='name of environment to apply, must be specified unless --cluster is')
    apply.set_defaults(func=apply_env)

    # diff = s.add_parser('diff', help='show diffs between running and git release')
    # diff.add_argument('--releases', '-r', help='which releases to diff', action='append')
    # diff.add_argument('env', nargs='?', help='name of environment')
    # diff.set_defaults(func=diff_release)

    poll_p = s.add_parser('poll',
                          help='poll for new images in registries')
    poll_p.add_argument('--patch', '-p', action='store_true',
                        help='patch releases.yaml files with updated version')
    poll_p.add_argument('--releases', '-r', action='append',
                        help='poll this specific release')
    poll_p.add_argument('--image', '-i',
                        help='poll releases using this image')
    poll_p.add_argument('--cluster', '-c',
                        help='poll all releases in this cluster')
    poll_p.add_argument('env', nargs='?',
                        help='name of environment to poll')
    poll_p.set_defaults(func=poll_registries)

    dump_p = s.add_parser('dump',
                          help='dump commands for one or all environments')
    dump_p.add_argument('env', nargs='?',
                        help='name of environment to dump')
    dump_p.set_defaults(func=dump_env)

    list_p = s.add_parser('list',
                          help='list clusters, environments or releases')
    list_p.add_argument('kind', choices=['env', 'release', 'cluster'],
                        help='what to list')
    list_p.set_defaults(func=list_kind)

    indent_p = s.add_parser('indent',
                            help='canonically indent YAML files')
    indent_p.add_argument('files', nargs='+',
                          help='file[s] to indent')
    indent_p.set_defaults(func=indent_file)

    observe = s.add_parser('observe',
                           help='observe a new image version')
    observe.add_argument('--image', '-i', metavar='IMAGE:TAG',
                         help='the image, including tag')
    observe.add_argument('--patch', action='store_true', default=False,
                         help='patch release files with updated tags')
    observe.add_argument('--submit-pr', action='store_true', default=False,
                         help='submit a pull request with the updated tags')
    observe.set_defaults(func=observe_new_image)

    completion_p = s.add_parser('completion',
                                help='print shell completion script')
    completion_p.set_defaults(func=print_completion, prog=prog)

    j2y = s.add_parser('json2yaml', help='JSON to YAML conversion utility (stdin/stdout)')
    j2y.set_defaults(func=json2yaml)

    init = s.add_parser('init', help='Initialize credentials and contexts')
    init.add_argument('--cluster',
                      help='Initialize contexts for all environments in a cluster')
    init.add_argument('--dry-run', '-n', action='store_true',
                      help='print commands instead of running them')
    init.add_argument('env_name', metavar='ENV', nargs='?',
                      help='environment to initialize')
    init.add_argument('--contexts-only', action='store_true',
                      help='initialize contexts only, assuming that cluster credentials are set up')
    init.add_argument('--gitlab', action='store_true',
                      help='grab kube config from GitLab environment')
    init.set_defaults(func=init_contexts)

    use = s.add_parser('use',
                       help='switch kube context to the specified environment')
    use.add_argument('env', metavar='ENV',
                     help='environment name')
    use.set_defaults(func=use_env_context)

    lint = s.add_parser('lint',
                        help='inspect the contents of a release, exits with non-0 if there are issues')
    lint.add_argument('--cluster',
                      help='Lint all environments in a cluster')
    lint.add_argument('env_name', metavar='ENV', nargs='?',
                      help='environment name')
    lint.set_defaults(func=lint_environment)

    return p


def verbose_log_level(v):
    if v == 0:
        return logging.WARNING
    if v == 1:
        return logging.INFO
    return logging.DEBUG


def print_completion(prog, **kwargs):
    shell = os.path.basename(os.getenv('SHELL'))
    if shell == 'bash' or shell == 'tcsh':
        sys.stdout.write(argcomplete.shellcode(prog, shell=shell))


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
