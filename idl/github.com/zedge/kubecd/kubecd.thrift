namespace py kubecd.gen_py


struct NameFromRef {
    1: optional string clusterParam;
}

struct GceAddressValueRef {
    1: optional string name;
    2: optional NameFromRef nameFrom;
    3: optional bool isGlobal; // if false, use zone from environment
}

union GceValueRef {
    1: optional GceAddressValueRef address;
}

struct ChartValueRef {
    1: optional GceValueRef gceResource;
}

struct ChartValue {
    1: optional string key;
    2: optional string value;
    3: optional ChartValueRef valueFrom;
}

struct GithubTrigger {
    1: optional string repository;
    2: optional string branch;
    3: optional string tagMatching;
}

struct ImageTrigger {
    1: optional string tagValue = "image.tag";
    2: optional string repoValue = "image.repository";
    3: optional string repoPrefixValue = "image.prefix";
    4: optional string track = "Newest"; // PatchLevel, MinorVersion, MajorVersion, Newest
}

union DeploymentTrigger {
    1: optional GithubTrigger github;
    2: optional ImageTrigger image;
}

struct Chart {
    1: optional string reference;
    2: optional string dir;
    3: optional string version;
}

struct Release {
    1: optional string name;
    2: optional Chart chart;
    3: optional string valuesFile;
    4: optional list<ChartValue> values;
    5: optional DeploymentTrigger trigger;
    6: optional list<DeploymentTrigger> triggers;
    7: optional bool skipDefaultValues;
}

struct KubernetesResourceRef {
    1: optional string kind;
    2: optional string name;  // optionally with a "namespace/" prefix
}

struct Releases {
    1: optional list<string> resourceFiles;
    2: optional list<Release> releases;
    3: optional list<KubernetesResourceRef> resourceDependencies;
}

struct GkeProvider {
    1: optional string project;
    2: optional string clusterName;
    3: optional string zone;
}

struct MinikubeProvider {
}

union Provider {
    1: optional GkeProvider gke;
    2: optional MinikubeProvider minikube;
}

struct ClusterParameter {
    1: optional string name;
    2: optional string value;
}

struct Cluster {
    1: optional string name;
    2: optional Provider provider;
    3: optional list<ClusterParameter> parameters;
}

struct Environment {
    1: optional string name;

    2: optional string clusterName;

    3: optional string kubeNamespace;

    /** a list of `releases.yaml` files */
    4: optional list<string> releasesFiles;

    /** default helm values file for the environment */
    5: optional string defaultValuesFile;

    6: optional list<ChartValue> defaultValues;
    
}

struct HelmRepo {
    1: optional string name;
    2: optional string url;
    3: optional string caFile;
    4: optional string certFile;
    5: optional string keyFile;
}

struct KubeCDConfig {
    1: optional list<Cluster> clusters;
    2: optional list<Environment> environments;
    3: optional list<HelmRepo> helmRepos;
}
