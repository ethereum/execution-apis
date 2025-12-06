import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';

export default function Home() {
  return (
    <Layout>
      <main style={{padding: '2rem', textAlign: 'center'}}>
        <h1>Ethereum Execution APIs</h1>
        <p>JSON-RPC API specification for Ethereum execution clients</p>
        <Link to="/api">View API Reference →</Link>
      </main>
    </Layout>
  );
}