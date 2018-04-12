from setuptools import setup, find_packages
from pipenv.project import Project
from pipenv.utils import convert_deps_to_pip

pfile = Project(chdir=False).parsed_pipfile
requirements = convert_deps_to_pip(pfile['packages'], r=False)
test_requirements = convert_deps_to_pip(pfile['dev-packages'], r=False)
setup(
    name='kubecd',
    version='0.1',
    description='Kubernetes Continuous Deployment and Inventory Tool',
    url='http://github.com/zedge/kubecd',
    author='Stig Bakken',
    author_email='stig@zedge.net',
    license='Apache 2',
    packages=find_packages(),
    install_requires=requirements,
    tests_require=test_requirements,
    entry_points={
        'console_scripts': [
            'kcd=zedge_kcd.cli:main'
        ]
    }
)
