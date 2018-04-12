namespace py kubecd.gen_py


struct GceAddressValueRef {
    1: required string name;
    2: optional bool isGlobal; // if false, use zone from environment
}

union GceValueRef {
    1: optional GceAddressValueRef address;
}

struct ChartValueRef {
    1: optional GceValueRef gceResource;
}

struct ChartValue {
    1: required string key;
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
    2: optional string repoValue = "image.value";
    3: optional string track = "Newest"; // PatchLevel, MinorVersion, MajorVersion, Newest
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
    1: required string name;
    2: optional Chart chart;
    3: optional string valuesFile;
    4: optional list<ChartValue> values;
    5: optional DeploymentTrigger trigger;
}

struct ResourceRef {
    1: required string kind;
    2: required string name;  // optionally with a "namespace/" prefix
}

struct Releases {
    1: optional list<string> resourceFiles;
    2: optional list<Release> releases;
    3: optional list<ResourceRef> resourceDependencies;
}

struct GkeProvider {
    1: required string project;
    2: required string clusterName;
    3: optional string zone;
}

union Provider {
    1: optional GkeProvider gke;
}

struct Cluster {
    1: required string name;
    2: required Provider provider;
}

struct Environment {
    1: required string name;

    2: required string clusterName;

    3: required string kubeNamespace;

    /** a list of `releases.yaml` files */
    4: optional list<string> releasesFiles;

    /** default helm values file for the environment */
    5: optional string defaultValuesFile;

    6: optional list<ChartValue> defaultValues;
    
}

struct Environments {
    1: required list<Cluster> clusters;
    2: required list<Environment> environments;
}
