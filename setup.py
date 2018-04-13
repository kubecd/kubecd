import os

from setuptools import setup, find_packages

module_path = os.path.join(os.path.dirname(__file__), 'kubecd', '__init__.py')
version_line = [line for line in open(module_path)
                if line.startswith('__version__')][0]

__version__ = version_line.split('__version__ = ')[-1][1:][:-2]

setup(
    name='kubecd',
    version=__version__,
    description='Kubernetes Continuous Deployment and Inventory Tool',
    url='http://github.com/zedge/kubecd',
    author='Stig Bakken',
    author_email='stig@zedge.net',
    license='Apache 2',
    packages=find_packages(),
    install_requires=['thrift', 'ruamel.yaml', 'semver', 'requests', 'click'],
    tests_require=['pytest', 'pytest-cov'],
    entry_points={
        'console_scripts': [
            'kcd=kubecd.cli:main'
        ]
    }
)
