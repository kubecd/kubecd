import abc
from typing import List

from .utils import kube_context
from .gen_py.ttypes import Environment, Cluster
from .gen_py import ttypes


class BaseClusterProvider(metaclass=abc.ABCMeta):
    def __init__(self, cluster: ttypes.Cluster):
        self.cluster = cluster

    @abc.abstractmethod
    def cluster_init_commands(self) -> List[List[str]]:
        pass

    @abc.abstractmethod
    def cluster_name(self) -> str:
        pass

    @abc.abstractmethod
    def user_name(self) -> str:
        pass

    def context_init_commands(self, env: Environment) -> List[List[str]]:
        context_name = kube_context(env.name)
        return [[
            'kubectl', 'config', 'set-context', context_name,
            '--cluster', self.cluster_name(),
            '--user', self.user_name(),
            '--namespace', env.kubeNamespace
        ]]


class GkeClusterProvider(BaseClusterProvider):
    def cluster_init_commands(self) -> List[List[str]]:
        return [[
            'gcloud', 'container', 'clusters', 'get-credentials',
            '--project', self.cluster.provider.gke.project,
            '--zone', self.cluster.provider.gke.zone,
            self.cluster.provider.gke.clusterName
        ]]

    def cluster_name(self) -> str:
        return self.cluster.provider.gke.clusterName

    def user_name(self) -> str:
        return self.cluster_name()


class AksClusterProvider(BaseClusterProvider):
    def cluster_init_commands(self) -> List[List[str]]:
        return [[
            'az', 'aks', 'get-credentials',
            '--resource-group', self.cluster.provider.aks.resourceGroup,
            '--name', self.cluster.provider.aks.clusterName
        ]]

    def cluster_name(self) -> str:
        return self.cluster.provider.aks.clusterName

    def user_name(self) -> str:
        return 'clusterUser_{aks.resourceGroup}_{aks.clusterName}'.format(aks=self.cluster.provider.aks)


class MinikubeClusterProvider(BaseClusterProvider):
    def cluster_init_commands(self) -> List[List[str]]:
        return [[
            'gcloud', 'container', 'clusters', 'get-credentials',
            '--project', self.cluster.provider.gke.project,
            '--zone', self.cluster.provider.gke.zone,
            self.cluster.provider.gke.clusterName
        ]]

    def cluster_name(self) -> str:
        return 'minikube'

    def user_name(self) -> str:
        return 'minikube'


def get_cluster_provider(cluster: Cluster) -> BaseClusterProvider:
    if cluster.provider.gke:
        return GkeClusterProvider(cluster)
    elif cluster.provider.minikube:
        return MinikubeClusterProvider(cluster)
    elif cluster.provider.aks:
        return AksClusterProvider(cluster)
    raise ValueError('missing or unknown cluster provider')


def cluster_init_commands(cluster: Cluster, env: Environment) -> List[List[str]]:
    """
    :param cluster:  cluster object for the environment
    :param env:      which environment we should generate init command[s] for
    :return: a list of commands, each command is an argv list
    """
    cluster_provider = get_cluster_provider(cluster)
    commands = cluster_provider.cluster_init_commands()
    commands.extend(cluster_provider.context_init_commands(env))
    return commands
