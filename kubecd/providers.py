from typing import List

from .utils import kube_context
from .gen_py.ttypes import Environment, Cluster
from .gen_py import ttypes


def generate_environment_init_command(cluster: Cluster, env: Environment, dry_run: bool = False) -> List[List[str]]:
    """
    :param cluster:  cluster object for the environment
    :param env:      which environment we should generate init command[s] for
    :param dry_run:  if True, do not generate mutating commands
    :return: a list of commands, each command is an argv list
    """
    if cluster.provider.gke:
        return gke_init_commands(env, cluster.provider.gke)
    elif cluster.provider.minikube:
        return minikube_init_commands(env, cluster.provider.minikube)
    elif cluster.provider.aks:
         return aks_init_commands(env, cluster.provider.aks)
    return []


def set_context_command(env: ttypes.Environment, cluster_name: str, user_name: str, namespace: str) -> List[str]:
    context_name = kube_context(env.name)
    return ['kubectl', 'config', 'set-context', context_name,
            '--cluster', cluster_name,
            '--user', user_name,
            '--namespace', namespace]


def gke_init_commands(env: ttypes.Environment, gke: ttypes.GkeProvider) -> List[List[str]]:
    cluster_name = 'gke_{project}_{zone}_{cluster}'.format(project=gke.project,
                                                           zone=gke.zone,
                                                           cluster=gke.clusterName)
    user_name = cluster_name
    return [
        ['gcloud', 'container', 'clusters', 'get-credentials',
         '--project', gke.project, '--zone', gke.zone, gke.clusterName],
        set_context_command(env, cluster_name, user_name, env.kubeNamespace)
    ]


def minikube_init_commands(env: ttypes.Environment, minikube: ttypes.MinikubeProvider) -> List[List[str]]:
    return [
        set_context_command(env, 'minikube', 'minikube', env.kubeNamespace)
    ]


def aks_init_commands(env: ttypes.Environment, aks: ttypes.AksProvider) -> List[List[str]]:
    # clusterUser_myResourceGroup_myAKSCluster
    user_name = 'clusterUser_{aks.resourceGroup}_{aks.clusterName}'.format(aks=aks)
    # az aks get-credentials --resource-group myResourceGroup --name myAKSCluster
    return [
        ['az', 'aks', 'get-credentials', '--resource-group', aks.resourceGroup, '--name', aks.clusterName],
        set_context_command(env, aks.clusterName, user_name, env.kubeNamespace)
    ]
