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

INSERT INTO achievements (
  id,
  code,
  title,
  description,
  category,
  threshold,
  is_active,
  created_at,
  updated_at
)
VALUES
  (
    '77777777-7777-7777-7777-777777777777',
    'streak_7_days',
    '7 Hari Konsisten',
    'Pertahankan streak keberhasilan selama 7 hari berturut-turut.',
    'streak_milestone',
    7,
    true,
    now(),
    now()
  ),
  (
    '88888888-8888-8888-8888-888888888888',
    'checkin_20_of_30',
    'Konsisten Check-in',
    'Lakukan check-in berhasil minimal 20 kali dalam 30 hari terakhir.',
    'checkin_consistency',
    20,
    true,
    now(),
    now()
  ),
  (
    '99999999-9999-9999-9999-999999999999',
    'journal_15_of_30',
    'Jurnal Reflektif',
    'Tulis jurnal minimal 15 kali dalam 30 hari terakhir.',
    'journal_consistency',
    15,
    true,
    now(),
    now()
  ),
  (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'relapse_recovery_3',
    'Bangkit Setelah Relapse',
    'Raih 3 check-in berhasil setelah relapse terakhir.',
    'relapse_recovery',
    3,
    true,
    now(),
    now()
  ),
  (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'community_participation_10',
    'Aktif di Komunitas',
    'Capai total 10 interaksi komunitas (post, komentar, atau like).',
    'community_participation',
    10,
    true,
    now(),
    now()
  ),
  (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'onboarding_complete',
    'Onboarding Selesai',
    'Selesaikan onboarding profil pemulihan.',
    'onboarding_completion',
    1,
    true,
    now(),
    now()
  )
ON CONFLICT (code) DO NOTHING;
