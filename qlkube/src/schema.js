const {GraphQLString, GraphQLSchema, GraphQLObjectType, GraphQLBoolean} = require('graphql');
const {mergeSchemas} = require('graphql-tools');
const {createGraphQlSchema} = require('oasgraph');

exports.createSchema = async (oas, kubeApiUrl, token) => {
    let baseSchema = await oasToGraphQlSchema(oas, kubeApiUrl, token)
    return decorateSchema(baseSchema)
};

async function oasToGraphQlSchema(oas, kubeApiUrl, token) {
    const {schema} = await createGraphQlSchema(oas, {
        baseUrl: kubeApiUrl,
        viewer: false,
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
    });
    return schema
}

// adds an 'all' type to the schema
function decorateSchema(baseSchema) {
    //TODO: extract query type names to env vars
    const allType = new GraphQLObjectType({
        name: 'all',
        description: 'All kube resources.',
        fields: {
            services: createType({name: "services", allNamespaceQueryName: "listCoreV1ServiceForAllNamespaces", namespacedQueryName: "ioK8sApiCoreV1ServiceList", baseSchema}),
            deployments: createType({name: "deployments", allNamespaceQueryName: "ioK8sApiAppsV1DeploymentList", namespacedQueryName: "listAppsV1NamespacedDeployment", baseSchema}),
            pods: createType({name: "pods", allNamespaceQueryName: "listCoreV1PodForAllNamespaces", namespacedQueryName: "ioK8sApiCoreV1PodList", baseSchema}),
            daemonSets: createType({name: "daemonSets", allNamespaceQueryName: "ioK8sApiAppsV1DaemonSetList", namespacedQueryName: "listAppsV1NamespacedDaemonSet", baseSchema}),
            replicaSets: createType({name: "replicaSets", allNamespaceQueryName: "listAppsV1ReplicaSetForAllNamespaces", namespacedQueryName: "ioK8sApiAppsV1ReplicaSetList", baseSchema}),
            statefulSets: createType({name: "statefulSets", allNamespaceQueryName: "listAppsV1StatefulSetForAllNamespaces", namespacedQueryName: "ioK8sApiAppsV1StatefulSetList", baseSchema}),
            jobs: createType({name: "jobs", allNamespaceQueryName: "ioK8sApiBatchV1JobList", namespacedQueryName: "listBatchV1NamespacedJob", baseSchema}),
            cronJobs: createType({name: "cronJobs", allNamespaceQueryName: "ioK8sApiBatchV1beta1CronJobList", namespacedQueryName: "listBatchV1beta1NamespacedCronJob", baseSchema}),
            namespaces: createNamespaceType({name: "namespaces", allNamespaceQueryName: "ioK8sApiCoreV1NamespaceList", baseSchema}),
            configmaps: createType({name: "configmaps", allNamespaceQueryName: "ioK8sApiCoreV1ConfigMapList", namespacedQueryName: "listCoreV1NamespacedConfigMap", baseSchema}),
        }
    });

    const schema = new GraphQLSchema({
        query: new GraphQLObjectType({
            name: 'query',
            fields: {
                all: {
                    type: allType,
                    args: {
                        fieldSelector: {type: GraphQLString},
                        namespace: {type: GraphQLString},
                        labelSelector: {type: GraphQLString}
                    },
                    resolve(_, args) {
                        return args // pass above args down to child resolvers as 'parent' object (args fields won't be accessible in served object)
                    }
                }
            }
        })
    });

    const schemas = [baseSchema, schema];
    return mergeSchemas({schemas});
}

// creates a type that lifts the return type and resolve methods from the existing generated schema methods
function createType({name, allNamespaceQueryName, namespacedQueryName, baseSchema}) {
    const allNamespaceQueryType = baseSchema.getQueryType().getFields()[allNamespaceQueryName];
    const namespacedQueryType = baseSchema.getQueryType().getFields()[namespacedQueryName];
    return {
        type: allNamespaceQueryType.type, //return type is the same for both namespaced and all-namespace variants.
        description: `All ${name} in all namespaces.`,
        resolve(parent, args, context, info) {
            return parent.namespace ?
                namespacedQueryType.resolve(parent, parent, context, info) : // 'parent' in this case is just the args passed to the 'all' parent, which is what we want to use as args here
                allNamespaceQueryType.resolve(parent, parent, context, info);
        }
    }
}

// Namespace has to be handled slightly differently to pods, deployments etc. as the singular 'namespaced' query type (`ioK8sApiCoreV1Namespace`)
// returns a different (singular) type rather than a `IoK8sApiCoreV1NamespaceList` - we must therefore instead manually filter
// the list to ensure parity.
function createNamespaceType({allNamespaceQueryName, baseSchema}) {
    const allNamespaceQueryType = baseSchema.getQueryType().getFields()[allNamespaceQueryName];
    return {
        type: allNamespaceQueryType.type,
        description: `All namespaces.`,
        async resolve(parent, args, context, info) {
            const namespacesProm = allNamespaceQueryType.resolve(parent, parent, context, info);
            if (parent.namespace) { // if 'namespace' argument provided to `all` query, then filter items in namespace list
                const namespaces = await namespacesProm;
                namespaces.items = namespaces.items.filter(ns => ns.metadata.name === parent.namespace);
                return namespaces
            }
            return namespacesProm
        }
    }
}

// TODO: all by namespace
// TODO: all by label
// TODO: serve raw schema

