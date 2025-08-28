// server.js (schema-agnostic, env-configurable)
import 'dotenv/config';
import express from 'express';
import cors from 'cors';
import { z } from 'zod';
import { ethers } from 'ethers';
import serverless from 'serverless-http';


// ───────────────────────────────────────────────────────────────────────────────
// ENV
// ───────────────────────────────────────────────────────────────────────────────
const {
  PORT = 4000,

  // Chain + token
  TOKEN_ADDRESS,
  RPC_URL,
  CHAIN_ID,

  // The Graph
  SUBGRAPH_URL,
  SUBGRAPH_API_KEY, // optional (Studio API key)
  SUBGRAPH_AUTH_TOKEN,
  EXPLORER_BASE_URL = 'https://etherscan.io',

  // ==== Schema config (set these to match YOUR subgraph) ====
  // Name of the entity *list field* (lowerCamelCase plural). Example: donations, transfers, transferEvents
  ENTITY_NAME = 'donations',

  // Address variable type in your schema: "Bytes" or "String"
  ADDR_VAR_TYPE = 'Bytes', // or "String"

  // Field names inside the entity
  FIELD_TX = 'txHash',           // e.g. txHash | transactionHash | hash
  FIELD_TS = 'timestamp',        // e.g. timestamp | blockTimestamp | ts
  FIELD_FROM = 'from',           // e.g. from | fromAddress | sender
  FIELD_TO = 'to',               // e.g. to | toAddress | recipient
  FIELD_AMOUNT = 'amount',       // e.g. amount | value | rawValue

  // Token field: can be scalar (address string/bytes) or an object (relation)
  TOKEN_IS_OBJECT = 'false',     // 'true' or 'false' (string)
  FIELD_TOKEN = 'token',         // name of the token field
  TOKEN_OBJECT_SUBFIELD = 'address', // used only if TOKEN_IS_OBJECT==='true'
} = process.env;

if (!TOKEN_ADDRESS || !RPC_URL || !CHAIN_ID || !SUBGRAPH_URL) {
  console.error(
    'Missing env. Please set TOKEN_ADDRESS, RPC_URL, CHAIN_ID, SUBGRAPH_URL. Also set schema config if defaults do not match your subgraph.'
  );
  process.exit(1);
}

const TOKEN_IS_OBJ = String(TOKEN_IS_OBJECT).toLowerCase() === 'true';

// ───────────────────────────────────────────────────────────────────────────────
// App & Ethers
// ───────────────────────────────────────────────────────────────────────────────
const app = express();
app.use(cors());
app.use(express.json());

const provider = new ethers.JsonRpcProvider(RPC_URL);
const tokenAbi = [
  'function name() view returns (string)',
  'function symbol() view returns (string)',
  'function decimals() view returns (uint8)',
];
const token = new ethers.Contract(TOKEN_ADDRESS, tokenAbi, provider);

let cachedMeta = null;
async function loadTokenMeta() {
  if (cachedMeta) return cachedMeta;
  const [name, symbol, decimals] = await Promise.all([
    token.name(),
    token.symbol(),
    token.decimals(),
  ]);
  cachedMeta = { name, symbol, decimals: Number(decimals) };
  return cachedMeta;
}

// ───────────────────────────────────────────────────────────────────────────────
// Graph helper
// ───────────────────────────────────────────────────────────────────────────────
async function querySubgraph(query, variables) {
  const headers = { 'content-type': 'application/json' };
  if (process.env.SUBGRAPH_API_KEY) headers['x-api-key'] = process.env.SUBGRAPH_API_KEY;
  if (process.env.SUBGRAPH_AUTH_TOKEN) headers.Authorization = `Bearer ${process.env.SUBGRAPH_AUTH_TOKEN}`;
  //if (SUBGRAPH_API_KEY) headers['x-api-key'] = SUBGRAPH_API_KEY;
  //if (SUBGRAPH_AUTH_TOKEN) headers.Authorization = `Bearer ${SUBGRAPH_AUTH_TOKEN}`;

  const resp = await fetch(SUBGRAPH_URL, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query, variables }),
  });

  const text = await resp.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    throw new Error(`Subgraph returned non-JSON (status ${resp.status}): ${text}`);
  }

  if (!resp.ok || json.errors) {
    throw new Error(
      `Subgraph error (HTTP ${resp.status}): ${JSON.stringify(json.errors || json)}`
    );
  }
  if (!json.data) {
    throw new Error(`Subgraph returned empty data (HTTP ${resp.status}): ${text}`);
  }
  return json.data;
}

// ───────────────────────────────────────────────────────────────────────────────
// Zod schemas
// ───────────────────────────────────────────────────────────────────────────────
const PassbookBody = z.object({
  address: z.string().min(1),
  page: z.coerce.number().int().min(1).default(1),
  pageSize: z.coerce.number().int().min(1).max(100).default(50),
});

// ───────────────────────────────────────────────────────────────────────────────
// Utilities to build GraphQL
// ───────────────────────────────────────────────────────────────────────────────
function buildSelection() {
  // Build selection set for one entity row, adapting token scalar vs object.
  const base = [
    'id',
    FIELD_TX,
    FIELD_TS,
    FIELD_FROM,
    FIELD_TO,
    FIELD_AMOUNT,
  ].filter(Boolean);

  if (FIELD_TOKEN) {
    if (TOKEN_IS_OBJ) {
      return `${base.join(' ')} ${FIELD_TOKEN} { ${TOKEN_OBJECT_SUBFIELD} }`;
    } else {
      return `${base.join(' ')} ${FIELD_TOKEN}`;
    }
  }
  return base.join(' ');
}

function buildQuery(addrVarType, entityName) {
  const selection = buildSelection();
  // Use aliases "in" and "out" for two lists.
  return /* GraphQL */ `
    query ($addr: ${addrVarType}!, $first: Int!, $skip: Int!) {
      in: ${entityName}(
        where: { ${FIELD_TO}: $addr }
        orderBy: ${FIELD_TS}
        orderDirection: desc
        first: $first
        skip: $skip
      ) { ${selection} }
      out: ${entityName}(
        where: { ${FIELD_FROM}: $addr }
        orderBy: ${FIELD_TS}
        orderDirection: desc
        first: $first
        skip: $skip
      ) { ${selection} }
    }
  `;
}

const GQL_BYTES = buildQuery('Bytes', ENTITY_NAME);
const GQL_STRING = buildQuery('String', ENTITY_NAME);

// ───────────────────────────────────────────────────────────────────────────────
// Routes
// ───────────────────────────────────────────────────────────────────────────────
app.get('/health', (_req, res) => res.json({ ok: true }));

app.get('/config', async (_req, res) => {
  try {
    const meta = await loadTokenMeta();
    res.json({
      chainId: Number(CHAIN_ID),
      token: {
        address: TOKEN_ADDRESS,
        ...meta,
      },
      explorerBaseUrl: EXPLORER_BASE_URL,
      // Expose schema config for quick sanity checks
      schema: {
        entity: ENTITY_NAME,
        addrVarType: ADDR_VAR_TYPE,
        fields: {
          tx: FIELD_TX, ts: FIELD_TS, from: FIELD_FROM, to: FIELD_TO, amount: FIELD_AMOUNT,
          token: FIELD_TOKEN, tokenIsObject: TOKEN_IS_OBJ, tokenSubfield: TOKEN_OBJECT_SUBFIELD
        }
      }
    });
  } catch (e) {
    res.status(500).json({ error: 'Failed to load config', details: String(e) });
  }
});

app.post('/passbook', async (req, res) => {
  try {
    const { address, page, pageSize } = PassbookBody.parse(req.body);
    const addr = address.toLowerCase();
    const skip = (page - 1) * pageSize;

    let queryUsed = '';
    let data;
    try {
      queryUsed = 'Bytes';
      data = await querySubgraph(GQL_BYTES, { addr, first: pageSize, skip });
    } catch (eBytes) {
      // If your schema uses String for addresses
      queryUsed = 'String';
      data = await querySubgraph(GQL_STRING, { addr, first: pageSize, skip });
    }

    const meta = await loadTokenMeta();

    // Guard against missing arrays
    const arrIn = Array.isArray(data?.in) ? data.in : [];
    const arrOut = Array.isArray(data?.out) ? data.out : [];

    // Normalize rows to our output shape
    const readField = (row, name) => row?.[name];
    const readToken = (row) => {
      if (!FIELD_TOKEN) return null;
      if (TOKEN_IS_OBJ) return row?.[FIELD_TOKEN]?.[TOKEN_OBJECT_SUBFIELD] ?? null;
      return row?.[FIELD_TOKEN] ?? null;
    };

    const mergedMap = new Map();
    for (const d of [...arrIn, ...arrOut]) {
      // Use id if present; else fall back to txHash+index if available
      const id = d?.id ?? `${readField(d, FIELD_TX)}:${readField(d, 'logIndex') ?? ''}`;
      if (!id) continue;
      mergedMap.set(id, d);
    }

    const normalized = [...mergedMap.values()].map((d) => {
      const ts = Number(readField(d, FIELD_TS));
      const from = String(readField(d, FIELD_FROM) || '').toLowerCase();
      const to = String(readField(d, FIELD_TO) || '').toLowerCase();
      const tokenVal = readToken(d);
      const rawAmount = String(readField(d, FIELD_AMOUNT) || '0');
      const txHash = readField(d, FIELD_TX);

      const isIn = to === addr;
      return {
        id: d.id || `${txHash}:${ts}`,
        txHash,
        direction: isIn ? 'IN' : 'OUT',
        counterparty: isIn ? from : to,
        rawAmount,
        amount: ethers.formatUnits(rawAmount || '0', meta.decimals ?? 18),
        token: tokenVal,
        timestamp: ts,
      };
    });

    normalized.sort((a, b) => b.timestamp - a.timestamp);

    res.json({
      address,
      page,
      pageSize,
      count: normalized.length,
      chainId: Number(CHAIN_ID),
      explorerBaseUrl: EXPLORER_BASE_URL,
      token: { address: TOKEN_ADDRESS, ...meta },
      transfers: normalized,
      debug: {
        queryVariantTried: queryUsed,
        entity: ENTITY_NAME
      }
    });
  } catch (e) {
    console.error('Passbook error →', e);
    res.status(500).json({
      error: 'Subgraph query failed',
      details: e.message || String(e),
      hint: 'Check your schema env (ENTITY_NAME, FIELD_* names, ADDR_VAR_TYPE, TOKEN_*). Hit GET /config to see what is currently set.'
    });
  }
});

// ───────────────────────────────────────────────────────────────────────────────
// add this simple home & favicon routes (optional but helpful)
app.get('/', (_req, res) => res.json({ ok: true, service: 'passbook-api' }));
app.get('/favicon.ico', (_req, res) => res.status(204).end());

// Export the Express app directly for Vercel
export default app;

// Local dev server
if (!process.env.VERCEL) {
  app.listen(Number(PORT), () => {
    console.log(`Passbook API running on http://localhost:${PORT}`);
  });
}
