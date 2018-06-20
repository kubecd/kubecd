from .. import semver as sut


def test_normalize():
    assert '1.1.0' == sut.normalize('v1.1.0')
    assert '1.1.0' == sut.normalize('1.1.0')


def test_parse():
    assert [1, 1, 0, (), ()] == list((sut.parse('v1.1.0')))
    assert [1, 1, 0, ('2',), ('build3',)] == list((sut.parse('v1.1.0-2+build3')))
    assert [2, 0, 0, (), ()] == list(sut.parse('2.0'))
