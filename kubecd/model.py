from collections import defaultdict
from os import path
from typing import List, Union, Dict

from . import helm
from .gen_py import ttypes
from .providers import generate_environment_init_command
from .thriftutils import load_yaml_with_schema
from .utils import resolve_file_path

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
        self._image_index = None
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
    def image_index(self) -> Dict[str, List[Release]]:
        index = defaultdict(list)
        if self._image_index is None:
            for env in self._environments:
                for rel in env.all_releases:
                    values = helm.get_resolved_values(rel, env, skip_value_from=True)
                    if rel.triggers:
                        for t in rel.triggers:
                            if t.image:
                                repo = helm.lookup_value(t.image.repoValue, values)
                                if repo is None:
                                    break
                                prefix = helm.lookup_value(t.image.repoPrefixValue, values)
                                if prefix is not None:
                                    repo = prefix + repo
                                index[repo].append(rel)
            self._image_index = index
        return self._image_index

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


# def generate_helm_command(rel: Release, env: Environment, release_file: str, dry_run: bool = False) -> str:
#     return ' '.join([shlex.quote(arg) for arg in generate_helm_command_argv(rel, env, release_file, dry_run)])

def releases_subscribing_to_image(image: str):
    pass


def config() -> KubeCDConfig:
    return _config

