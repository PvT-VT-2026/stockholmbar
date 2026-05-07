const express = require('express');
const app = express();
app.use(express.json());
app.use(express.static('public'));

const submissions = [
  { id: 'abc-1', submitted_by: 'user-1', category: 'venue', status: 'pending',
    payload: { name: 'Foobar', street: 'Testgatan 1', city: 'Stockholm' },
    created_at: new Date().toISOString() },
  { id: 'abc-2', submitted_by: 'user-2', category: 'venue', status: 'pending',
    payload: { name: 'Baz', street: 'Exempelvägen 5', city: 'Göteborg' },
    created_at: new Date().toISOString() },
];

let queue = [...submissions];

app.get('/admin/submission/list', (req, res) => {
  const status = req.query.status;
  res.json(status ? submissions.filter(s => s.status === status) : submissions);
});

app.get('/admin/submission/next', (req, res) => {
  const next = queue.shift();
  if (!next) return res.status(204).end();
  res.json(next);
});

app.post('/admin/submission/:id/accept', (req, res) => {
  console.log('ACCEPT', req.params.id);
  res.json({ ok: true });
});

app.post('/admin/submission/:id/reject', (req, res) => {
  console.log('REJECT', req.params.id);
  res.json({ ok: true });
});

// Returnerar en platshållarbild
app.get('/admin/submission/:id/image', (req, res) => {
  res.redirect(`https://picsum.photos/seed/${req.params.id}/400/300`);
});

app.listen(8081, () => console.log('Listening on :8081'));