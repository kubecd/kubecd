import json
import os
from typing import Dict
from unittest.mock import patch

import pytest

from .. import updates as sut

TEST_DOCKER_HUB_TAGS_URL = 'https://registry.hub.docker.com/v2/repositories/confluentinc/cp-kafka/tags?page_size=256'
TEST_REGISTRY_TAGS_URL = 'https://docker.trd.zedge.net:5001/v2/zedge/frontend/tags/list'
TEST_REGISTRY_MANIFESTS_URL_PREFIX = 'https://docker.trd.zedge.net:5001/v2/zedge/frontend/manifests/'


def local_file(file_name: str):
    with open(os.path.join(os.path.dirname(__file__), file_name)) as f:
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
        l = len(TEST_REGISTRY_MANIFESTS_URL_PREFIX)
        tag = args[0][l:]
        return MockResponse(200, local_file('registry-v2-manifests-{tag}.json'.format(tag=tag)))
    else:
        pytest.fail("unexpected requests.get url: {url}".format(url=args[0]))


@patch('subprocess.check_output')
def test_get_tags_for_gcr_image(subprocess_mock, gcr_airflow_tags, gcr_expected_tags):
    subprocess_mock.return_value = gcr_airflow_tags.encode('utf-8')
    tags = sut.get_tags_for_gcr_image('us.gcr.io', 'zedge-prod/airflow')
    assert tags == gcr_expected_tags


@patch('requests.get', side_effect=mocked_requests_get)
def test_get_tags_for_docker_hub_image(mock_get, docker_hub_expected_tags):
    tags = sut.get_tags_for_docker_hub_image('confluentinc/cp-kafka')
    assert tags == docker_hub_expected_tags


@patch('requests.get', side_effect=mocked_requests_get)
def test_get_tags_for_docker_v2_registry(mock_get, docker_registry_v2_expected_tags):
    tags = sut.get_tags_for_docker_v2_registry('docker.trd.zedge.net:5001', 'zedge/frontend')
    print(tags)
    assert tags == docker_registry_v2_expected_tags


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
        '3.0.0': 1469720357,
        '3.0.1': 1476815495,
        '3.0.1-1': 1473197814,
        '3.0.1-2': 1475184713,
        '3.0.1-3': 1476815492,
        '3.1.0': 1478706768,
        '3.1.0-1': 1478706765,
        '3.1.1': 1479250079,
        '3.1.1-1': 1479250075,
        '3.1.2': 1488478302,
        '3.1.2-1': 1485814059,
        '3.1.2-2': 1488478296,
        '3.2.0': 1490319276,
        '3.2.0-1': 1488587303,
        '3.2.0-2': 1488922605,
        '3.2.0-3': 1490303778,
        '3.2.0-4': 1490319273,
        '3.2.1': 1494621855,
        '3.2.1-4': 1493823869,
        '3.2.1-5': 1493915178,
        '3.2.1-6': 1494621853,
        '3.2.2': 1498516047,
        '3.2.2-1': 1498516044,
        '3.3.0': 1501525447,
        '3.3.0-1': 1501525444,
        '3.3.1': 1510597295,
        '4.0.0': 1515004297,
        '4.0.0-2': 1515004295,
        '4.0.0-3': 1518727312,
        'e618702': 1469641365,
        'latest': 1518727314
    }


@pytest.fixture
def docker_registry_v2_expected_tags() -> Dict[str, int]:
    return {
        'test':   1423818295,
        'latest': 1437461276,
    }
