from collections import defaultdict
from os import path
from typing import List, Union, Dict

import logging

from . import helm
from .gen_py import ttypes
from .providers import cluster_init_commands
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
        issues = self._sanity_check()
        if len(issues) > 0:
            raise ValueError('release {} issues found: {}'.format(self.name, ', '.join(issues)))
        self._compat_update()

    def _compat_update(self):
        if self.trigger:
            self.triggers = [self.trigger]
            self.trigger = None

    def _sanity_check(self):
        issues = []
        if self.chart is None and self.resourceFiles is None:
            issues.append('must define either "chart" or "resourceFiles"')
        if self.chart is not None and self.resourceFiles is not None:
            issues.append('must define only one of "chart" or "resourceFiles"')
        if self.chart is not None:
            if self.chart.reference is not None and self.chart.version is None:
                issues.append('must have a chart.version'.format(self.chart.name))
        if self.trigger and self.triggers:
            issues.append('must define only one of "trigger" or "triggers"')
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
        issues = self._sanity_check()
        if len(issues) > 0:
            raise ValueError('Issues found:\n\t{}'.format('\n\t'.join(issues)))
        for rel_file in self.releasesFiles:
            abs_file = resolve_file_path(rel_file, relative_to_file=self._from_file)
            releases = load_yaml_with_schema(abs_file, ttypes.Releases)
            if releases.resourceFiles is not None:
                logging.warning('%s: resourceFiles at the top level are deprecated, move them to the a release',
                                abs_file)
                abs_files = [resolve_file_path(x, relative_to_file=abs_file) for x in releases.resourceFiles]
                self._all_resource_files.extend(abs_files)
            for rel in releases.releases:
                self._all_releases.append(Release(rel, from_file=abs_file))

    def _sanity_check(self):
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

    def init_commands(self):
        cluster = get_cluster(self.clusterName)
        return cluster_init_commands(cluster, self)

    @property
    def cluster(self):
        return get_cluster(self.clusterName)


class Cluster(ttypes.Cluster):
    def __init__(self, data: ttypes.Cluster, from_file: str):
        self._from_file = from_file
        super(Cluster, self).__init__(**data.__dict__)

    def __hash__(self) -> int:
        return hash(self.name)

    @property
    def from_file(self):
        return self._from_file


class KubecdConfig(ttypes.KubecdConfig):
    # _environments: List[Environment]
    # _index: Dict[str, Environment] = {}
    # _from_file: str

    def __init__(self, data: ttypes.KubecdConfig, from_file: str):
        self._environments = [Environment(x, from_file) for x in data.environments]
        self._clusters = [Cluster(x, from_file) for x in data.clusters]
        self._from_file = from_file
        self._env_index = {}
        self._cluster_index = {}
        self._image_index = None
        super(KubecdConfig, self).__init__(**data.__dict__)
        issues = self._sanity_check()
        if len(issues) > 0:
            raise ValueError('Issues found:\n\t{}'.format('\n\t'.join(issues)))
        for env in self._environments:
            self._env_index[env.name] = env
        for cluster in self._clusters:
            self._cluster_index[cluster.name] = cluster

    def __iter__(self):
        return iter(self._environments)

    def _sanity_check(self):
        counts = defaultdict(int)
        issues = []
        for env in self._environments:
            if env.name in counts:
                issues.append('duplicate environment name: {}'.format(env.name))
            counts[env.name] += 1
        return issues

    def all_clusters(self) -> List[Cluster]:
        return list(self._cluster_index.values())

    def all_environments(self) -> List[Environment]:
        return self._environments

    def get_environment(self, name: str) -> Environment:
        return self._env_index[name]

    def get_cluster(self, name: str) -> Cluster:
        return self._cluster_index[name]

    def has_cluster(self, name: str) -> bool:
        return name in self._cluster_index

    def environments_in_cluster(self, cluster_name: str) -> List[Environment]:
        envs = []
        for env in self.all_environments():
            if env.clusterName == cluster_name:
                envs.append(env)
        return envs

    def init_commands(self) -> List[List[str]]:
        commands = []
        if self.helmRepos:
            commands.extend(helm.repo_setup_commands(self.helmRepos))
        return commands

    @property
    def image_index(self) -> Dict[str, List[Release]]:
        index = defaultdict(list)
        if self._image_index is None:
            logging.debug('building image index...')
            for env in self._environments:
                logging.debug(' - Environment %s', env.name)
                for rel in env.all_releases:
                    logging.debug('    - Release %s', rel.name)
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


def load(file_name: str) -> KubecdConfig:
    global _config
    _config = KubecdConfig(load_yaml_with_schema(file_name, ttypes.KubecdConfig), from_file=file_name)
    return _config


def all_environments() -> List[Environment]:
    return [x for x in config()]


def all_clusters() -> List[Cluster]:
    return list(set([e.cluster for e in all_environments()]))


def get_cluster(cluster_name: str) -> Cluster:
    return config().get_cluster(cluster_name)


def get_environment(env_name: str) -> Environment:
    return config().get_environment(env_name)


def environments_in_cluster(cluster_name: str) -> List[Environment]:
    envs = []
    for env in all_environments():
        if env.clusterName == cluster_name:
            envs.append(env)
    return envs


def config() -> KubecdConfig:
    if _config is None:
        raise ConfigError('environments not yet loaded')
    return _config
