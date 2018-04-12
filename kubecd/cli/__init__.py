import subprocess
import sys
from typing import List

import click
import yaml

from kubecd.updates import find_updates_for_env
from .. import environments


@click.group()
@click.option('--environments-file', '-f',
              envvar='KCD_ENVIRONMENTS',
              type=click.Path(exists=True),
              help='Your environments.yaml file',
              default='environments.yaml')
@click.pass_context
def main(ctx, environments_file):
    ctx.obj = {
        'environments_file': environments_file
    }


@main.group()
def env():
    pass


def load_envs(file_name: str) -> List[environments.Environment]:
    try:
        environments.load(file_name)
        return environments.as_list()
    except yaml.error.YAMLError as e:
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


@env.command(name='apply')
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


@env.command(name='list')
@click.pass_context
def list_envs(ctx):
    for environment in environments.load(ctx.obj['environments_file']):
        click.echo(environment.name)


@env.command(name='dump')
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


@main.command(name='check-updates')
@click.argument('ENV_NAME', required=False)
@click.pass_context
def check_updates(ctx, env_name=None):
    for environment in one_or_more_envs(env_name, ctx.obj['environments_file']):
        updates = find_updates_for_env(environment)
        for release in updates:
            for image_repo in updates[release]:
                click.echo('\timage "{image}" tag {tag_from} -> {tag_to}'.format(
                    env=environment.name,
                    release=release,
                    image=image_repo,
                    tag_from=updates[release][image_repo][0],
                    tag_to=updates[release][image_repo][1],
                ))


if __name__ == '__main__':
    main()
