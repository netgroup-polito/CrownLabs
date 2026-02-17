const { ApolloServer, expressMiddleware } = require('@apollo/server');
const { ApolloServerPluginDrainHttpServer } = require('@apollo/server/plugin/drainHttpServer');
const compression = require('compression');
const dotenv = require('dotenv');
const express = require('express');
const fs = require('fs').promises;
const { printSchema } = require('graphql');
const { createServer } = require('http');
const { WebSocketServer } = require('ws');
const { useServer } = require('graphql-ws/lib/use/ws');
const cors = require('cors');
const repl = require('repl');
const { setupSubscriptions } = require('./decorateSubscription');
const { getOpenApiSpec } = require('./oas');
const { decorateOpenapi } = require('./decorateOpenapi');
const { createSchema, oasToGraphQlSchema, joinSchemas } = require('./schema');
const { subscriptions } = require('./subscriptions');
const { kinformer } = require('./informer');
const { getBearerToken, logger } = require('./utils');
const { schema: nodeLabelsSchema } = require('./nodeLabels');

dotenv.config();

async function main() {
  const inCluster = process.env.IN_CLUSTER !== 'false';
  logger.info({ inCluster }, 'cluster mode configured');

  const kubeApiUrl = inCluster
    ? 'https://kubernetes.default.svc'
    : 'http://localhost:8001';
  const inClusterToken = inCluster
    ? await fs.readFile(
      '/var/run/secrets/kubernetes.io/serviceaccount/token',
      'utf8',
    )
    : '';
  const registryUrl = inCluster ? 'http://cloudimg-registry' : 'http://localhost:8002';

  const oas = await getOpenApiSpec(kubeApiUrl, ['openapi/v2'], inClusterToken);
  const targetOas = decorateOpenapi(oas);

  let kubeSchema = await createSchema(targetOas, kubeApiUrl, inClusterToken);
  kubeSchema = setupSubscriptions(subscriptions, kubeSchema);

  global.kubeSchema = kubeSchema;

  const schemas = [{ schema: kubeSchema }, { schema: nodeLabelsSchema }];

  try {
    const registryOas = await getOpenApiSpec(registryUrl, ['docs/openapi.json']);
    const registrySchema = (await oasToGraphQlSchema(registryOas, registryUrl, null, true)).schema;
    schemas.push({ schema: registrySchema, prefix: 'reg' });
    global.registrySchema = registrySchema;
  } catch (error) {
    logger.warn({ error }, 'Registry OAS fetch failure');
  }

  const schema = joinSchemas(schemas);

  const app = express();
  const httpServer = createServer(app);

  app.use(compression());

  app.get('/schema', (req, res) => {
    res.setHeader('content-type', 'text/plain');
    res.send(
      printSchema(schema)
        .split('\n')
        .filter((l) => l.trim() !== '_') // remove empty values from enums
        .join('\n'),
    );
  });
  app.get('/healthz', (req, res) => {
    res.sendStatus(200);
  });
  app.get('/', (req, res) => {
    res.setHeader('content-type', 'text/html');
    res.send(`<html><head><title>CrownLabs GraphQL Playground</title></head><body>
      <div></div>
      <script src="https://embeddable-sandbox.cdn.apollographql.com/_latest/embeddable-sandbox.umd.production.min.js"></script> 
      <script>new window.EmbeddedSandbox({ target: 'div', initialEndpoint: document.location.href, initialSubscriptionEndpoint: document.location.href+'subscription' });</script>
      <style>html, body, body > div {height: 100%;body: 100%;margin: 0;padding: 0}</style>
      </body></html>`);
  });

  let serverCleanup;

  const server = new ApolloServer({
    schema,
    plugins: [
      ApolloServerPluginDrainHttpServer({ httpServer }),
      {
        async serverWillStart() {
          return {
            async drainServer() {
              if (serverCleanup) {
                await serverCleanup.dispose();
              }
            },
          };
        },
      },
    ],
    formatError: (error) => {
      try {
        const msgs = [...(error.message.match(/(\d+)\s*-(.*)$/) || [])];
        if (msgs.length !== 3) return error;
        let msg = JSON.parse(msgs[2]);
        try {
          msg = JSON.parse(msg);
        } catch (_) { /* empty: file silently */ }
        return {
          ...error,
          message: msg.message,
          extensions: {
            ...error.extensions,
            k8s: msg,
            http: msgs[1],
          },
        };
      } catch (e) {
        logger.error(null, 'Cannot parse error response');
        // eslint-disable-next-line no-console
        console.error(e);
        return error;
      }
    },

    context: async ({ _req, connection }) => {
      if (connection) {
        const { token } = connection.context;
        return { token };
      }
      return null;
    },
  });

  // Creating the WebSocket server
  const wsServer = new WebSocketServer({
    // This is the `httpServer` we created in a previous step.
    server: httpServer,
    // Pass a different path here if app.use
    // serves expressMiddleware at a different path
    path: '/subscription',
  });

  // Hand in the schema we just created and have the
  // WebSocketServer start listening.
  serverCleanup = useServer({
    schema,
    onConnect: ({ connectionParams }, _webSocket, _context) => {
      const token = getBearerToken(connectionParams);
      if (!token) {
        logger.info('Incoming WS connection without token');
        return false;
      }
      try {
        const { preferred_username: decodedUser } = JSON.parse(atob(token.split('.')[1])); // JWT data part
        logger.info({ decodedUser }, 'Incoming WS connection');
      } catch (err) {
        logger.warn({ error: err.message }, 'Incoming WS connection - Token parsing failed');
      }
      return { token };
    },
    onDisconnect: (_webSocket, _context) => {
      logger.info('WS disconnected');
    },
    context: ({ connectionParams }, _msg, _args) => {
      const token = getBearerToken(connectionParams);
      return { token };
    },
  }, wsServer);

  await server.start();

  app.use(
    '/',
    cors(),
    express.json(),
    expressMiddleware(server, {
      context: async ({ req }) => {
        const token = getBearerToken(req.headers);
        return { token };
      },
    }),
  );

  const PORT = process.env.CROWNLABS_QLKUBE_PORT || 8080;

  httpServer.listen({ port: PORT }, () => {
    logger.info({
      url: `http://localhost:${PORT}/`,
      subscriptions: `ws://localhost:${PORT}/subscription`,
    }, '🚀 Server ready');
    repl.start('> ');
  });

  /**
   * Making informer for watching resources.
   */
  subscriptions.forEach(kinformer);
}

main().catch((e) => {
  logger.error({ error: e.message }, 'failed to start qlkube server');
  // eslint-disable-next-line no-console
  console.error(e);
});
