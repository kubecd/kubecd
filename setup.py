import os

from setuptools import setup, find_packages, Command

module_path = os.path.join(os.path.dirname(__file__), 'kubecd', '__init__.py')
version_line = [line for line in open(module_path)
                if line.startswith('__version__')][0]

__version__ = version_line.split('__version__ = ')[-1][1:][:-2]


class ReleaseCommand(Command):
    """ Run my command.
    """
    description = 'tag a new release'
    version = None
    force = False
    user_options = [
        ('version=', 'v', 'new version'),
        ('force', 'f', 'force push version tag')
    ]

    def initialize_options(self):
        self.version = None
        self.force = False

    def finalize_options(self):
        if self.version is None:
            self.version = __version__

    def run(self):
        import subprocess
        tag = 'v' + self.version
        push_cmd = ["git", "push", "--tags"]
        if self.force:
            push_cmd.append("-f")
            print("Untagging %s ..." % tag)
            if subprocess.call(["git", "tag", "-d", tag]) != 0:
                raise Exception("'git untag' command failed")
        print("Tagging %s ..." % tag)
        if subprocess.call(["git", "tag", tag]) != 0:
            raise Exception("'git tag' command failed")
        print("Pushing tags ...")
        if subprocess.call(push_cmd) != 0:
            raise Exception("'git push --tags' command failed")


setup(
    name='kubecd',
    version=__version__,
    cmdclass={
        'release': ReleaseCommand
    },
    description='Kubernetes Continuous Deployment and Inventory Tool',
    url='http://github.com/zedge/kubecd',
    author='Stig Bakken',
    author_email='stig@zedge.net',
    license='Apache 2',
    packages=find_packages(),
    python_requires='>=3.5.0',
    install_requires=[
        'thrift==0.11.0',
        'ruamel.yaml~=0.15.37',
        'semantic_version~=2.6.0',
        'requests~=2.18.4',
        'argcomplete~=1.9.4',
        'argparse~=1.4.0',
        'blessings',
        'python-dateutil'
    ],
    tests_require=['pytest', 'pytest-cov'],
    entry_points={
        'console_scripts': [
            'kcd=kubecd.cli:main'
        ]
    }
)
