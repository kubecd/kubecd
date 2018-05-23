import subprocess
import time
from typing import List


def pr_from_files(files: List[str], message: str):
    branch_name = 'br%d' % time.time()
    subprocess.run(['git', 'checkout', '-b', branch_name], check=True)
    git_add = ['git', 'add']
    git_add.extend(files)
    # print('git_add_cmd:', git_add)
    subprocess.run(git_add, check=True)
    subprocess.run(['git', 'commit', '-m', message], check=True)
    subprocess.run(['git', 'push', 'origin', branch_name], check=True)
    subprocess.run(['hub', 'pull-request', '-m', message], check=True)
