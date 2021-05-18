const fs = require('fs').promises;
const { createServer } = require('http');
const express = require('express');
const { ApolloServer } = require('apollo-server-express');
const compression = require('compression');
const { createSchema } = require('./schema');
const { kwatch } = require('./watch.js');
const { setupSubscriptions } = require('./decorateSubscription.js');
const { subscriptions } = require('./subscriptions.js');
const getOpenApiSpec = require('./oas');
const { printSchema } = require('graphql');
const logger = require('pino')({ useLevelLabels: true });
const dotenv = require('dotenv');

dotenv.config();

main().catch(e =>
  logger.error({ error: e.stack }, 'failed to start qlkube server')
);

async function main() {
  const inCluster = process.env.IN_CLUSTER !== 'false';
  logger.info({ inCluster }, 'cluster mode configured');
  const kubeApiUrl = inCluster
    ? 'https://kubernetes.default.svc'
    : 'http://localhost:8001';
  const token = inCluster
    ? await fs.readFile(
        '/var/run/secrets/kubernetes.io/serviceaccount/token',
        'utf8'
      )
    : '';

  const oas = await getOpenApiSpec(kubeApiUrl, token);
  let schema = await createSchema(oas, kubeApiUrl, token);

  try {
    schema = setupSubscriptions(subscriptions, schema);
  } catch (e) {
    console.error(e);
    process.exit(1);
  }

  const server = new ApolloServer({
    schema,
    subscriptions: {
      path: '/subscription',
      onConnect: (connectionParams, webSocket, context) => {
        console.log('Connected!');
      },
      onDisconnect: (webSocket, context) => {
        console.log('Disconnected!');
      },
    },

    context: ({ req, connection }) => {
      if (connection) {
        return {};
      } else {
        if (req.headers.authorization && req.headers.authorization.length > 0) {
          const strs = req.headers.authorization.split(' ');
          var user = {};
          user.token = strs[1];
          return user;
        }
      }
    },
  });

  const app = express();
  app.use(compression());
  app.get('/schema', (req, res) => {
    res.setHeader('content-type', 'text/plain');
    res.send(printSchema(schema));
  });
  app.get('/healthz', (req, res) => {
    res.sendStatus(200);
  });
  server.applyMiddleware({
    app,
    path: '/',
  });
  const httpServer = createServer(app);
  server.installSubscriptionHandlers(httpServer);

  const PORT = process.env.CROWNLABS_QLKUBE_PORT || 8080;

  httpServer.listen({ port: PORT }, () => {
    console.log(
      `🚀 Server ready at http://localhost:${PORT}${server.graphqlPath}`
    );
    console.log(
      `🚀 Subscriptions ready at ws://localhost:${PORT}${server.subscriptionsPath}`
    );
  });

  try {
    subscriptions.forEach(sub => {
      kwatch(sub.resource, sub.type);
    });
  } catch (e) {
    console.error(e);
    process.exit(1);
  }
}
