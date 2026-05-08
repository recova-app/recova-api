INSERT INTO education_contents (
  id,
  title,
  description,
  url,
  thumbnail_url,
  category,
  is_active,
  published_at
)
VALUES
  (
    '11111111-1111-1111-1111-111111111111',
    'Memahami Trigger dan Rutinitas',
    'Dasar mengenali pemicu harian dan membentuk respon yang lebih sehat.',
    'https://recova.app/education/memahami-trigger-dan-rutinitas',
    NULL,
    'mindset',
    true,
    now()
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'Teknik Grounding 5-4-3-2-1',
    'Latihan sederhana untuk kembali fokus saat dorongan muncul.',
    'https://recova.app/education/teknik-grounding-5-4-3-2-1',
    NULL,
    'coping',
    true,
    now()
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO daily_motivations (id, content, is_active, created_at)
VALUES
  (
    '33333333-3333-3333-3333-333333333333',
    'Satu keputusan sehat hari ini tetap berarti besar.',
    true,
    now()
  ),
  (
    '44444444-4444-4444-4444-444444444444',
    'Kemajuan kecil yang konsisten lebih kuat dari niat sesaat.',
    true,
    now()
  )
ON CONFLICT (content) DO NOTHING;

INSERT INTO daily_challenges (id, content, is_active, created_at)
VALUES
  (
    '55555555-5555-5555-5555-555555555555',
    'Catat satu pemicu utama hari ini dan rencana responnya.',
    true,
    now()
  ),
  (
    '66666666-6666-6666-6666-666666666666',
    'Lakukan jeda 60 detik sebelum bereaksi saat dorongan muncul.',
    true,
    now()
  )
ON CONFLICT (content) DO NOTHING;
