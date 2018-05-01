import ruamel.yaml
import re
import subprocess
from collections import defaultdict
from typing import List, Union

from .thriftutils import load_yaml_with_schema
from .utils import resolve_file_path
from .providers import generate_environment_init_command
from .gen_py import ttypes
from os import path

_config = None


class ConfigError(BaseException):
    pass


class Release(ttypes.Release):
    # _from_file: str

    def __init__(self, data: ttypes.Release, from_file: str):
        self._from_file = from_file
        super(Release, self).__init__(**data.__dict__)
        issues = self.sanity_check()
        if len(issues) > 0:
            raise ValueError('release {} issues found: {}'.format(self.name, ', '.join(issues)))

    def sanity_check(self):
        issues = []
        if self.chart.reference is not None and self.chart.version is None:
            issues.append('must have a chart.version'.format(self.chart.name))
        if self.trigger and self.triggers:
            issues.append('must define only one of "trigger" or "triggers"')
        if self.trigger:
            self.triggers = [self.trigger]
            self.trigger = None
        return issues

    def get_resolved_values(self, for_env: ttypes.Environment) -> dict:
        # 1. get values from chart if in a local dir
        # 2. merge with valuesFile:
        # 3. merge with values:
        values = {}
        if self.chart.dir:
            values_file = resolve_file_path(path.join(self.chart.dir, 'values.yaml'), relative_to_file=self._from_file)
            if path.exists(values_file):
                values = merge_values(from_dict=load_values_file(values_file), onto_dict=values)
        elif self.chart.reference:
            # "helm inspect" outputs a two-document yaml block, where the second is the parsed default values.yaml
            cmd = ['helm', 'inspect', self.chart.reference, '--version', self.chart.version]
            output = subprocess.check_output(cmd).decode('utf-8').split("\n---\n")[1]
            chart_default_values = ruamel.yaml.safe_load(output)
            values = merge_values(from_dict=chart_default_values, onto_dict=values)
        if len(for_env.defaultValues) > 0:
            default_values = values_list_to_dict(for_env.defaultValues, for_env)
            values = merge_values(from_dict=default_values, onto_dict=values)
        if self.valuesFile:
            values_file = resolve_file_path(self.valuesFile, relative_to_file=self._from_file)
            values = merge_values(from_dict=load_values_file(values_file), onto_dict=values)
        if self.values:
            values = merge_values(from_dict=values_list_to_dict(self.values, for_env), onto_dict=values)
        return values

    @property
    def from_file(self):
        return self._from_file


class Environment(ttypes.Environment):
    # _from_file: str
    # _all_releases: List[Release]
    # _all_resource_files: List[str]

    def __init__(self, data: ttypes.Environment, from_file: str):
        self._from_file = path.abspath(from_file)
        self._all_releases = []
        self._all_resource_files = []
        super(Environment, self).__init__(**data.__dict__)
        issues = self.sanity_check()
        if len(issues) > 0:
            raise ValueError('Issues found:\n\t{}'.format('\n\t'.join(issues)))
        for rel_file in self.releasesFiles:
            abs_file = resolve_file_path(rel_file, relative_to_file=self._from_file)
            releases = load_yaml_with_schema(abs_file, ttypes.Releases)
            if releases.resourceFiles is not None:
                abs_files = [resolve_file_path(x, relative_to_file=abs_file) for x in releases.resourceFiles]
                self._all_resource_files.extend(abs_files)
            for rel in releases.releases:
                self._all_releases.append(Release(rel, from_file=abs_file))

    def sanity_check(self):
        counts = defaultdict(int)
        issues = []
        for env in self.all_releases:
            if env.name in counts:
                issues.append('duplicate environment name: {}'.format(env.name))
            counts[env.name] += 1
        return issues

    @property
    def all_releases(self) -> List[Release]:
        return self._all_releases

    def named_release(self, name: str) -> Union[Release, None]:
        for r in self._all_releases:
            if r.name == name:
                return r
        return None

    @property
    def all_resource_files(self):
        return self._all_resource_files

    def init_commands(self, dry_run=False):
        cluster = get_cluster(self.clusterName)
        return generate_environment_init_command(cluster, self, dry_run)

    def deploy_commands(self, dry_run=False, limit_to_release=None) -> List[List[str]]:
        commands = []
        if limit_to_release is None:
            for resource_file in self.all_resource_files:
                cmd = ['kubectl', '--context', 'env:{}'.format(self.name), 'apply']
                if dry_run:
                    cmd.append('--dry-run')
                cmd.extend(['-f', resource_file])
                commands.append(cmd)
        for release in self._all_releases:
            if not limit_to_release or release.name == limit_to_release:
                rel_file = release.from_file
                commands.append(generate_helm_command_argv(release, self, release_file=rel_file, dry_run=dry_run))
        return commands

    @property
    def cluster(self):
        return get_cluster(self.clusterName)


class Cluster(ttypes.Cluster):
    def __init__(self, data: ttypes.Cluster, from_file: str):
        self._from_file = from_file
        super(Cluster, self).__init__(**data.__dict__)

    @property
    def from_file(self):
        return self._from_file


class KubeCDConfig(ttypes.KubeCDConfig):
    # _environments: List[Environment]
    # _index: Dict[str, Environment] = {}
    # _from_file: str

    def __init__(self, data: ttypes.KubeCDConfig, from_file: str):
        self._environments = [Environment(x, from_file) for x in data.environments]
        self._clusters = [Cluster(x, from_file) for x in data.clusters]
        self._from_file = from_file
        self._env_index = {}
        self._cluster_index = {}
        super(KubeCDConfig, self).__init__(**data.__dict__)
        issues = self.sanity_check()
        if len(issues) > 0:
            raise ValueError('Issues found:\n\t{}'.format('\n\t'.join(issues)))
        for env in self._environments:
            self._env_index[env.name] = env
        for cluster in self._clusters:
            self._cluster_index[cluster.name] = cluster

    def __iter__(self):
        return iter(self._environments)

    def sanity_check(self):
        counts = defaultdict(int)
        issues = []
        for env in self._environments:
            if env.name in counts:
                issues.append('duplicate environment name: {}'.format(env.name))
            counts[env.name] += 1
        return issues

    def env_by_name(self, name: str) -> Environment:
        return self._env_index[name]

    def cluster_by_name(self, name: str) -> Cluster:
        return self._cluster_index[name]

    def init_commands(self) -> List[List[str]]:
        commands = []
        if self.helmRepos:
            for repo in self.helmRepos:
                commands.append(['helm', 'repo', 'add', repo.name, repo.url])
        return commands

    @property
    def from_file(self):
        return self._from_file


def load(file_name: str) -> KubeCDConfig:
    global _config
    _config = KubeCDConfig(load_yaml_with_schema(file_name, ttypes.KubeCDConfig), from_file=file_name)
    return _config


def as_list() -> List[Environment]:
    if _config is None:
        raise ConfigError('environments not yet loaded')
    return [x for x in _config]


def get_cluster(cluster_name: str) -> Cluster:
    if _config is None:
        raise ConfigError('environments not yet loaded')
    return _config.cluster_by_name(cluster_name)


def get_environment(env_name: str) -> Environment:
    if _config is None:
        raise ConfigError('environments not yet loaded')
    return _config.env_by_name(env_name)


def generate_helm_command_argv(rel: Release, env: Environment, release_file: str, dry_run: bool = False) -> List[str]:
    chart_arg = rel.chart.reference
    if chart_arg is None:
        chart_arg = resolve_file_path(rel.chart.dir, release_file)
        if not path.exists(chart_arg):
            raise ValueError('{}: release "{}" chart.dir "{}" does not exist'.format(
                release_file, rel.name, chart_arg))
    argv = ['helm', '--kube-context', 'env:' + env.name, 'upgrade', rel.name, chart_arg, '-i']
    if dry_run:
        argv.append('--dry-run')
    if env.defaultValuesFile:
        def_val_file = resolve_file_path(env.defaultValuesFile, release_file)
        argv.extend(['--values', def_val_file])
    if env.defaultValues:
        argv.append('--set')
        argv.append(','.join(['='.join(resolve_value(x, env)) for x in env.defaultValues]))
    if rel.valuesFile:
        val_file = resolve_file_path(rel.valuesFile, release_file)
        argv.extend(['--values', val_file])
    if rel.values:
        argv.append('--set')
        argv.append(','.join(['='.join(resolve_value(x, env)) for x in rel.values]))
    return argv


# def generate_helm_command(rel: Release, env: Environment, release_file: str, dry_run: bool = False) -> str:
#     return ' '.join([shlex.quote(arg) for arg in generate_helm_command_argv(rel, env, release_file, dry_run)])


def resolve_gce_address_value(address: ttypes.GceAddressValueRef, env: Environment):
    cmd = ['gcloud', 'compute', 'addresses', 'describe', address.name, '--format', 'value(address)']
    provider = env.cluster.provider
    cmd.extend(['--project', provider.gke.project])
    if address.isGlobal:
        cmd.append('--global')
    else:
        cmd.extend(['--region', re.sub(r'-[a-z]$', '', provider.gke.zone)])
    return subprocess.check_output(cmd).decode('utf-8').strip()


def resolve_value(value: ttypes.ChartValue, env: Environment) -> tuple:
    v = value.value
    if value.valueFrom:
        if value.valueFrom.gceResource:
            if value.valueFrom.gceResource.address:
                v = resolve_gce_address_value(value.valueFrom.gceResource.address, env)
    return value.key, v


def merge_values(from_dict: dict, onto_dict: dict) -> dict:
    for key, value in from_dict.items():
        if isinstance(value, dict):
            # get node or create one
            node = onto_dict.setdefault(key, {})
            if isinstance(node, dict):
                merge_values(value, node)
        else:
            onto_dict[key] = value
    return onto_dict


def load_values_file(file_name: str) -> dict:
    with open(file_name) as f:
        return ruamel.yaml.safe_load(f)


def values_list_to_dict(values: List[ttypes.ChartValue], env: Environment) -> dict:
    """
    Convert a list of ChartValue objects with keys on the form ``"foo.bar"``
    to a nested dictionary like ``{"foo": {"bar": ...}}``.
    :param values: list of ChartValue objects
    :param env: current environment
    :return:
    """
    def val_to_dict(key: List[str], value) -> dict:
        if len(key) == 1:
            return {key[0]: value}
        return {key[0]: val_to_dict(key[1:], value)}

    output = {}
    for value_obj in values:
        k, v = resolve_value(value_obj, env)
        output = merge_values(from_dict=val_to_dict(k.split('.'), v), onto_dict=output)
    return output


def lookup_value(key: Union[List[str], str], values: dict):
    if type(key) == str:
        key = key.split('.')
    if len(key) > 0 and len(values) > 0:
        if key[0] in values:
            if len(key) == 1:
                return values[key[0]]
            return lookup_value(key[1:], values[key[0]])
    return None


def key_is_in_values(key: Union[List[str], str], values: dict) -> bool:
    if type(key) == str:
        key = key.split('.')
    if len(key) > 0 and len(values) > 0:
        if len(key) == 1:
            return key[0] in values
        if key[0] in values:
            return key_is_in_values(key[1:], values[key[0]])
    return False


def releases_subscribing_to_image(image: str):
    pass


def config() -> KubeCDConfig:
    return _config

