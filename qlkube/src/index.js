const { ApolloServer } = require('apollo-server-express');
const compression = require('compression');
const dotenv = require('dotenv');
const express = require('express');
const fs = require('fs').promises;
const { printSchema } = require('graphql');
const { createServer } = require('http');
const logger = require('pino')({ useLevelLabels: true });
const { setupSubscriptions } = require('./decorateSubscription.js');
const getOpenApiSpec = require('./oas');
const { createSchema } = require('./schema');
const { subscriptions } = require('./subscriptions.js');
const { kwatch } = require('./watch.js');

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
    schema = setupSubscriptions(subscriptions, schema, kubeApiUrl);
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
        const token =
          connectionParams.authorization &&
          connectionParams.authorization.split(' ')[1];
        return token ? { token } : false;
      },
      onDisconnect: (webSocket, context) => {
        console.log('Disconnected!');
      },
    },

    context: ({ req, connection }) => {
      if (connection) {
        const token = connection.context.token;
        return { token };
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
      const resourceApi = `/${sub.api}${sub.group ? `/${sub.group}` : ''}/${
        sub.version
      }/${sub.resource}`;
      kwatch(resourceApi, sub.type);
    });
  } catch (e) {
    console.error(e);
    process.exit(1);
  }
}
