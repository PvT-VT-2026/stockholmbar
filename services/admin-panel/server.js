const express = require('express');
const app = express();
app.use(express.json());
app.use(express.static('public'));

const API_BASE  = process.env.API_BASE_URL;
const API_TOKEN = process.env.API_TOKEN;

if (!API_BASE || !API_TOKEN) {
  console.warn('API_BASE_URL och/eller API_TOKEN saknas.');
}

const HEADERS = {
  'Authorization': `Bearer ${API_TOKEN}`,
  'apikey': API_TOKEN,
  'Content-Type': 'application/json',
  'Prefer': 'return=representation',
};

async function supabase(res, path, { method = 'GET', body } = {}) {
  try {
    const upstream = await fetch(`${API_BASE}/rest/v1${path}`, {
      method,
      headers: HEADERS,
      ...(body ? { body: JSON.stringify(body) } : {}),
    });
 
    const text = await upstream.text();
 
    if (!upstream.ok) {
      console.error(`Upstream ${method} /rest/v1${path} → ${upstream.status}:`, text);
      return res.status(upstream.status).json({ error: text });
    }
 
    return res.json(text ? JSON.parse(text) : {});
  } catch (err) {
    console.error('Proxy-fel:', err);
    return res.status(502).json({ error: 'Kunde inte nå Supabase.' });
  }
}

app.get('/admin/submission/list', (req, res) => {
  const filter = req.query.status
    ? `?status=eq.${req.query.status}&order=created_at.asc`
    : '?order=created_at.asc';
  supabase(res, `/submission${filter}`);
});

app.get('/admin/submission/next', async (req, res) => {
  try {
    const upstream = await fetch(
      `${API_BASE}/rest/v1/submission?status=eq.pending&order=created_at.asc&limit=1`,
      { headers: HEADERS }
    );
    const text = await upstream.text();
    if (!upstream.ok) return res.status(upstream.status).json({ error: text });
 
    const rows = JSON.parse(text);
    if (!rows.length) return res.status(204).end();
    res.json(rows[0]);
  } catch (err) {
    res.status(502).json({ error: 'Kunde inte nå Supabase.' });
  }
});

app.post('/admin/submission/:id/accept', (req, res) => {
  supabase(res, `/submission?id=eq.${req.params.id}`, {
    method: 'PATCH',
    body: { status: 'accepted', reviewed_at: new Date().toISOString() },
  });
});

app.post('/admin/submission/:id/reject', (req, res) => {
  supabase(res, `/submission?id=eq.${req.params.id}`, {
    method: 'PATCH',
    body: { status: 'rejected', reviewed_at: new Date().toISOString() },
  });
});

app.get('/admin/submission/:id/image', async (req, res) => {
  const storageUrl = `${API_BASE}/storage/v1/object/public/submission/${req.params.id}`;
  try {
    const check = await fetch(storageUrl, { method: 'HEAD', headers: HEADERS });
    if (check.ok) return res.redirect(storageUrl);
  } catch (_) {}
  res.redirect(`https://picsum.photos/seed/${req.params.id}/400/300`);
});

app.listen(8081, () => console.log('Listening on :8081'));