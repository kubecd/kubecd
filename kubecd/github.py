import subprocess
import time
from typing import List


def pr_from_files(files: List[str], message: str):
    branch_name = 'br%d' % time.time()
    subprocess.call(['git', 'checkout', '-b', branch_name])
    git_add = ['git', 'add']
    git_add.extend(files)
    print('git_add_cmd:', git_add)
    subprocess.call(git_add)
    subprocess.call(['git', 'commit', '-m', message])
    subprocess.call(['git', 'push', 'origin', branch_name])
    subprocess.call(['hub', 'pull-request', '-m', message])
