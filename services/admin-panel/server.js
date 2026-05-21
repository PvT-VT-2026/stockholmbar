const express = require('express');
const app = express();
app.use(express.json());
app.use(express.static('public'));
 
const API_BASE     = process.env.API_BASE_URL;
const SUPABASE_URL = process.env.SUPABASE_URL;
const API_TOKEN    = process.env.API_TOKEN;
const ANON_KEY     = process.env.SUPABASE_ANON_KEY;

if (!API_BASE || !SUPABASE_URL || !API_TOKEN || !ANON_KEY) {
  console.warn('Saknas miljövariabler: API_BASE_URL, SUPABASE_URL, API_TOKEN, SUPABASE_ANON_KEY');
}

const HEADERS = {
  'Authorization': `Bearer ${API_TOKEN}`,
  'apikey': API_TOKEN,
  'Content-Type': 'application/json',
};

app.get('/api/config', (req, res) => {
  res.json({ supabaseUrl: SUPABASE_URL, supabaseAnonKey: ANON_KEY });
});

async function proxy(res, path, { method = 'GET', body } = {}) {
  try {
    const upstream = await fetch(`${API_BASE}${path}`, {
      method,
      headers: HEADERS,
      ...(body ? { body: JSON.stringify(body) } : {}),
    });

    const text = await upstream.text();

    if (!upstream.ok) {
      console.error(`Upstream ${method} ${path} → ${upstream.status}:`, text);
      return res.status(upstream.status).json({ error: text });
    }

    if (!text) return res.status(upstream.status).end();
    return res.json(JSON.parse(text));
  } catch (err) {
    console.error('Proxy-fel:', err);
    return res.status(502).json({ error: 'Kunde inte nå API.' });
  }
}

app.get('/admin/submission/list', (req, res) => {
  const qs = req.query.status ? `?status=${req.query.status}` : '';
  proxy(res, `/admin/submission/list${qs}`);
});

app.get('/admin/submission/next', (req, res) => {
  proxy(res, '/admin/submission/next');
});

app.post('/admin/submission/:id/accept', (req, res) => {
  proxy(res, `/admin/submission/${req.params.id}/accept`, { method: 'POST', body: req.body });
});

app.post('/admin/submission/:id/reject', (req, res) => {
  proxy(res, `/admin/submission/${req.params.id}/reject`, { method: 'POST', body: req.body });
});

app.get('/admin/submission/:id/image', async (req, res) => {
  const storageUrl = `${API_BASE}/admin/submission/${req.params.id}/image`;
  try {
    const check = await fetch(storageUrl, { method: 'HEAD', headers: HEADERS });
    if (check.ok) return res.redirect(storageUrl);
  } catch (_) {}
  res.redirect(`https://picsum.photos/seed/${req.params.id}/400/300`);
});

app.listen(8081, () => console.log('Listening on :8081'));