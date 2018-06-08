import json
import os
from typing import Dict
from unittest.mock import patch

import pytest

from .. import updates as sut

TEST_DOCKER_HUB_TAGS_URL = 'https://registry.hub.docker.com/v2/repositories/confluentinc/cp-kafka/tags?page_size=256'
TEST_REGISTRY_TAGS_URL = 'https://docker.example.net:5001/v2/example/application/tags/list'
TEST_REGISTRY_MANIFESTS_URL_PREFIX = 'https://docker.example.net:5001/v2/example/application/manifests/'


def local_file(file_name: str):
    with open(os.path.join(os.path.dirname(__file__), 'testdata', file_name)) as f:
        return f.read()


def test_parse_image_repo():
    assert sut.parse_image_repo('centos') == ('docker.io', 'centos')
    assert sut.parse_image_repo('gcr.io/foo') == ('gcr.io', 'foo')


def mocked_requests_get(*args, **kwargs):
    class MockResponse:
        def __init__(self, status, response):
            self.status = status
            self.response = response

        def json(self):
            return json.loads(self.response)

    if args[0] == TEST_DOCKER_HUB_TAGS_URL:
        return MockResponse(200, local_file('docker-hub-kafka-tags.json'))
    elif args[0] == TEST_REGISTRY_TAGS_URL:
        return MockResponse(200, local_file('registry-v2-tags-list.json'))
    elif args[0].startswith(TEST_REGISTRY_MANIFESTS_URL_PREFIX):
        ln = len(TEST_REGISTRY_MANIFESTS_URL_PREFIX)
        tag = args[0][ln:]
        return MockResponse(200, local_file('registry-v2-manifests-{tag}.json'.format(tag=tag)))
    else:
        pytest.fail("unexpected requests.get url: {url}".format(url=args[0]))


def test_parse_docker_timestamp():
    ts = sut.parse_docker_timestamp('1970-01-01T02:00:00Z')
    assert ts == 7200, 'ts is {ts}'.format(ts=ts)


@patch('subprocess.check_output')
def test_get_tags_for_gcr_image(subprocess_mock, gcr_airflow_tags, gcr_expected_tags):
    subprocess_mock.return_value = gcr_airflow_tags.encode('utf-8')
    tags = sut.get_tags_for_gcr_image('us.gcr.io', 'gcp-project/airflow')
    assert tags == gcr_expected_tags


@patch('requests.get', side_effect=mocked_requests_get)
def test_get_tags_for_docker_hub_image(mock_get, docker_hub_expected_tags):
    tags = sut.get_tags_for_docker_hub_image('confluentinc/cp-kafka')
    assert tags == docker_hub_expected_tags


@patch('requests.get', side_effect=mocked_requests_get)
def test_get_tags_for_docker_v2_registry(mock_get, docker_registry_v2_expected_tags):
    tags = sut.get_tags_for_docker_v2_registry('docker.example.net:5001', 'example/application')
    print(tags)
    assert tags == docker_registry_v2_expected_tags


def test_get_newest_matching_tag(docker_hub_expected_tags):
    ver = '3.1.1'
    ts = docker_hub_expected_tags[ver]
    patch_update = sut.get_newest_matching_tag(ver, docker_hub_expected_tags, 'PatchLevel', ts)
    assert patch_update == '3.1.2'
    minor_update = sut.get_newest_matching_tag(ver, docker_hub_expected_tags, 'MinorVersion', ts)
    assert minor_update == '3.3.1'
    major_update = sut.get_newest_matching_tag(ver, docker_hub_expected_tags, 'MajorVersion', ts)
    assert major_update == '4.0.0'
    tags = {'1.0.0': 0, '1.0.0-0': 0}
    assert '1.0.0' == sut.get_newest_matching_tag('0.9', tags, 'MajorVersion', 0)
    tags = {'0.9': 2, '1.0': 1}
    assert '0.9' == sut.get_newest_matching_tag('0.9', tags, 'Newest', 0)


@pytest.fixture
def gcr_airflow_tags():
    return local_file('gcr-airflow-tags.json')


@pytest.fixture
def gcr_expected_tags():
    return {
        '1.8': 1512057463,
        '1.8.2': 1513690388
    }


@pytest.fixture
def registry_v2_tags_list():
    return local_file('registry-v2-tags-list.json')


@pytest.fixture
def registry_v2_manifests_test():
    return local_file('registry-v2-manifests-test.json')


@pytest.fixture
def docker_hub_kafka_tags():
    return local_file('docker-hub-kafka-tags.json')


@pytest.fixture
def docker_hub_expected_tags() -> Dict[str, int]:
    return {
        '3.0.0': 1469727557,
        '3.0.1': 1476822695,
        '3.0.1-1': 1473205014,
        '3.0.1-2': 1475191913,
        '3.0.1-3': 1476822692,
        '3.1.0': 1478710368,
        '3.1.0-1': 1478710365,
        '3.1.1': 1479253679,
        '3.1.1-1': 1479253675,
        '3.1.2': 1488481902,
        '3.1.2-1': 1485817659,
        '3.1.2-2': 1488481896,
        '3.2.0': 1490322876,
        '3.2.0-1': 1488590903,
        '3.2.0-2': 1488926205,
        '3.2.0-3': 1490307378,
        '3.2.0-4': 1490322873,
        '3.2.1': 1494629055,
        '3.2.1-4': 1493831069,
        '3.2.1-5': 1493922378,
        '3.2.1-6': 1494629053,
        '3.2.2': 1498523247,
        '3.2.2-1': 1498523244,
        '3.3.0': 1501532647,
        '3.3.0-1': 1501532644,
        '3.3.1': 1510600895,
        '4.0.0': 1515007897,
        '4.0.0-2': 1515007895,
        '4.0.0-3': 1518730912,
        'e618702': 1469648565,
        'latest': 1518730914
    }


@pytest.fixture
def docker_registry_v2_expected_tags() -> Dict[str, int]:
    return {
        'test': 1423821895,
        'latest': 1437468476,
    }
