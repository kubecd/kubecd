from setuptools import setup, find_packages

setup(
    name='kubecd',
    version='0.3',
    description='Kubernetes Continuous Deployment and Inventory Tool',
    url='http://github.com/zedge/kubecd',
    author='Stig Bakken',
    author_email='stig@zedge.net',
    license='Apache 2',
    packages=find_packages(),
    install_requires=['pyyaml', 'thrift', 'ruamel.yaml', 'semver', 'requests', 'click'],
    tests_require=['pytest', 'pytest-cov'],
    entry_points={
        'console_scripts': [
            'kcd=kubecd.cli:main'
        ]
    }
)
