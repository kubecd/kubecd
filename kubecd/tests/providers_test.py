import pytest

from .. import providers as sut

GCP_PROJECT = 'gcp-project'
GCP_REGION = 'us-central1'
GCP_ZONE = 'us-central1-a'
GKE_CLUSTER_NAME = 'my-cluster'


def test_gke_regional(gke_regional_cluster):
    provider = sut.get_cluster_provider(gke_regional_cluster)
    assert [
               ['gcloud', 'container', 'clusters', 'get-credentials', '--project', GCP_PROJECT, '--region', GCP_REGION, GKE_CLUSTER_NAME]
           ] == provider.cluster_init_commands()
    assert 'gke_{project}_{region}_{cluster}'.format(project=GCP_PROJECT, region=GCP_REGION, cluster=GKE_CLUSTER_NAME) == provider.cluster_name()


def test_gke_zonal(gke_zonal_cluster):
    provider = sut.get_cluster_provider(gke_zonal_cluster)
    assert [
               ['gcloud', 'container', 'clusters', 'get-credentials', '--project', GCP_PROJECT, '--zone', GCP_ZONE, GKE_CLUSTER_NAME]
           ] == provider.cluster_init_commands()
    assert 'gke_{project}_{zone}_{cluster}'.format(project=GCP_PROJECT, zone=GCP_ZONE, cluster=GKE_CLUSTER_NAME) == provider.cluster_name()


@pytest.fixture
def gke_regional_cluster():
    return sut.ttypes.Cluster(
        name='regional-cluster',
        provider=sut.ttypes.Provider(
            gke=sut.ttypes.GkeProvider(project=GCP_PROJECT, clusterName=GKE_CLUSTER_NAME, region=GCP_REGION)))


@pytest.fixture
def gke_zonal_cluster():
    return sut.ttypes.Cluster(
        name='zonal-cluster',
        provider=sut.ttypes.Provider(
            gke=sut.ttypes.GkeProvider(project=GCP_PROJECT, clusterName=GKE_CLUSTER_NAME, zone=GCP_ZONE)))
