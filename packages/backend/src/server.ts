import Fastify from 'fastify';
import dotenv from 'dotenv';

// Load environment variables from the root .env.local file
dotenv.config({ path: '../../.env.local' });

// Initialize Fastify with logger enabled
const fastify = Fastify({
  logger: true,
});

// Health check endpoint
fastify.get('/health', async (request, reply) => {
  return { status: 'ok' };
});

// Start the server
const start = async () => {
  try {
    const port = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

    // Listen on 0.0.0.0 to ensure it accepts connections when running in containers/different network interfaces
    await fastify.listen({ port, host: '0.0.0.0' });
    fastify.log.info(`Server listening on port ${port}`);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
};

start();
