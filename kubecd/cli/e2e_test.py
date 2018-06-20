from os import path
from unittest.mock import patch, MagicMock

import pytest

from . import commands


@patch('builtins.print')
def test_dupe_ingress(mock_print: MagicMock):
    env_file = path.join(path.dirname(__file__), 'e2e', 'dupe-ingress', 'environments.yaml')
    with pytest.raises(commands.CliError):
        commands.lint_environment(environments_file=env_file, env_name='test')
    assert mock_print.call_args_list[0][0][0].startswith('ERROR: Ingress')
