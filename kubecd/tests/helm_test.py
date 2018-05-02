from .. import helm as sut


def test_lookup_value():
    assert 'foo' == sut.lookup_value('a.b', {'a': {'b': 'foo'}})
    assert 'foo' == sut.lookup_value(['a', 'b'], {'a': {'b': 'foo'}})
    assert sut.lookup_value(['b.a'], {'a': {'b': 'foo'}}) is None


def test_key_is_in_values():
    assert sut.key_is_in_values(['image', 'tag'], {'image': {'tag': 'foo'}})
    assert sut.key_is_in_values('image.tag', {'image': {'tag': 'foo'}})
    assert not sut.key_is_in_values(['image', 'tag'], {'imageTag': 'foo'})
    assert not sut.key_is_in_values(['image.tag'], {'imageTag': 'foo'})
