from typing import List

from .gen_py.ttypes import Environment, Cluster


def generate_environment_init_command(cluster: Cluster, env: Environment, dry_run: bool = False) -> List[List[str]]:
    """
    :param cluster:  cluster object for the environment
    :param env:      which environment we should generate init command[s] for
    :param dry_run:  if True, do not generate mutating commands
    :return: a list of commands, each command is an argv list
    """
    commands = []
    context_name = 'env:' + env.name
    if cluster.provider.gke:
        commands.append(['gcloud', 'container', 'clusters', 'get-credentials',
                         '--project', cluster.provider.gke.project,
                         '--zone', cluster.provider.gke.zone,
                         cluster.provider.gke.clusterName])
        cluster_name = 'gke_{project}_{zone}_{cluster}'.format(project=cluster.provider.gke.project,
                                                               zone=cluster.provider.gke.zone,
                                                               cluster=cluster.name)
        commands.append(['kubectl', 'config', 'set-context', context_name,
                         '--cluster', cluster_name,
                         '--user', cluster_name,
                         '--namespace', env.kubeNamespace])
    elif cluster.provider.minikube:
        cluster_name = 'minikube'
        commands.append(['kubectl', 'config', 'set-context', context_name,
                         '--cluster', cluster_name,
                         '--user', cluster_name,
                         '--namespace', env.kubeNamespace])
    return commands
