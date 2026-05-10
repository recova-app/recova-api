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
    'Memahami Pemicu dan Pola Harian',
    'Dasar mengenali pemicu harian dan menyusun respon yang lebih sehat.',
    'https://recova.app/education/memahami-pemicu-dan-pola-harian',
    NULL,
    'pola_pikir',
    true,
    now()
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'Teknik Grounding 5-4-3-2-1',
    'Latihan sederhana untuk mengembalikan fokus saat dorongan muncul.',
    'https://recova.app/education/teknik-grounding-5-4-3-2-1',
    NULL,
    'regulasi_emosi',
    true,
    now()
  ),
  (
    '12121212-1212-1212-1212-121212121212',
    'Menyusun Rencana Darurat 10 Menit',
    'Panduan membuat langkah cepat saat situasi terasa berat.',
    'https://recova.app/education/rencana-darurat-10-menit',
    NULL,
    'strategi_harian',
    true,
    now()
  ),
  (
    '13131313-1313-1313-1313-131313131313',
    'Jeda Napas untuk Menurunkan Impuls',
    'Teknik jeda napas untuk meredakan reaksi spontan.',
    'https://recova.app/education/jeda-napas-menurunkan-impuls',
    NULL,
    'regulasi_emosi',
    true,
    now()
  ),
  (
    '14141414-1414-1414-1414-141414141414',
    'Audit Lingkungan Pemulihan',
    'Cara menata lingkungan agar lebih mendukung proses pemulihan.',
    'https://recova.app/education/audit-lingkungan-pemulihan',
    NULL,
    'strategi_harian',
    true,
    now()
  ),
  (
    '15151515-1515-1515-1515-151515151515',
    'Menjaga Konsistensi di Hari Sibuk',
    'Kiat mempertahankan kebiasaan kecil saat jadwal padat.',
    'https://recova.app/education/konsistensi-di-hari-sibuk',
    NULL,
    'kebiasaan',
    true,
    now()
  ),
  (
    '16161616-1616-1616-1616-161616161616',
    'Mengenali Pikiran Otomatis Negatif',
    'Langkah praktis mengidentifikasi dan menantang pikiran otomatis negatif.',
    'https://recova.app/education/mengenali-pikiran-otomatis-negatif',
    NULL,
    'pola_pikir',
    true,
    now()
  ),
  (
    '17171717-1717-1717-1717-171717171717',
    'Refleksi Mingguan yang Efektif',
    'Template refleksi mingguan untuk melihat progres secara objektif.',
    'https://recova.app/education/refleksi-mingguan-efektif',
    NULL,
    'refleksi',
    true,
    now()
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO daily_motivations (id, content, is_active, created_at)
VALUES
  ('30000000-0000-0000-0000-000000000001', 'Satu langkah kecil hari ini tetap berarti untuk masa depanmu.', true, now()),
  ('30000000-0000-0000-0000-000000000002', 'Konsisten tidak harus sempurna, yang penting terus bergerak.', true, now()),
  ('30000000-0000-0000-0000-000000000003', 'Kamu tidak sendirian, minta dukungan adalah tindakan berani.', true, now()),
  ('30000000-0000-0000-0000-000000000004', 'Menunda dorongan selama beberapa menit sudah sebuah kemenangan.', true, now()),
  ('30000000-0000-0000-0000-000000000005', 'Fokus pada hari ini, bukan beban seluruh minggu.', true, now()),
  ('30000000-0000-0000-0000-000000000006', 'Progress yang tenang lebih kuat daripada semangat sesaat.', true, now()),
  ('30000000-0000-0000-0000-000000000007', 'Tarik napas, beri jeda, lalu pilih respon terbaikmu.', true, now()),
  ('30000000-0000-0000-0000-000000000008', 'Kamu berhak memulai ulang kapan pun kamu butuh.', true, now()),
  ('30000000-0000-0000-0000-000000000009', 'Setiap keputusan sehat hari ini memperkuat dirimu besok.', true, now()),
  ('30000000-0000-0000-0000-000000000010', 'Kemajuanmu nyata, meski tidak selalu terlihat besar.', true, now())
ON CONFLICT (content) DO NOTHING;

INSERT INTO daily_challenges (id, content, is_active, created_at)
VALUES
  ('40000000-0000-0000-0000-000000000001', 'Catat satu pemicu utama hari ini dan respon sehat yang kamu pilih.', true, now()),
  ('40000000-0000-0000-0000-000000000002', 'Lakukan jeda 60 detik sebelum merespons dorongan yang muncul.', true, now()),
  ('40000000-0000-0000-0000-000000000003', 'Tulis tiga hal kecil yang berhasil kamu jaga hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000004', 'Kurangi satu distraksi digital selama 30 menit untuk fokus pemulihan.', true, now()),
  ('40000000-0000-0000-0000-000000000005', 'Hubungi satu orang tepercaya untuk check-in singkat.', true, now()),
  ('40000000-0000-0000-0000-000000000006', 'Susun rencana tidur lebih awal minimal 30 menit dari biasanya.', true, now()),
  ('40000000-0000-0000-0000-000000000007', 'Lakukan jalan kaki ringan 10 menit sambil mengatur napas.', true, now()),
  ('40000000-0000-0000-0000-000000000008', 'Rapikan area kerja atau kamar selama 15 menit untuk reset pikiran.', true, now()),
  ('40000000-0000-0000-0000-000000000009', 'Ganti satu kebiasaan pemicu dengan aktivitas pengalih yang sehat.', true, now()),
  ('40000000-0000-0000-0000-000000000010', 'Tutup hari ini dengan jurnal refleksi singkat 5 kalimat.', true, now())
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
    'Pertahankan streak berhasil selama 7 hari berturut-turut.',
    'streak_milestone',
    7,
    true,
    now(),
    now()
  ),
  (
    '88888888-8888-8888-8888-888888888888',
    'checkin_20_of_30',
    'Check-in Konsisten',
    'Selesaikan minimal 20 check-in berhasil dalam 30 hari terakhir.',
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
    'Tulis minimal 15 jurnal dalam 30 hari terakhir.',
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
    'Capai 3 check-in berhasil setelah relapse terakhir.',
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
  ),
  (
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'streak_30_days',
    '30 Hari Beruntun',
    'Pertahankan streak berhasil selama 30 hari berturut-turut.',
    'streak_milestone',
    30,
    true,
    now(),
    now()
  ),
  (
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    'checkin_60_of_90',
    'Ketahanan 90 Hari',
    'Selesaikan minimal 60 check-in berhasil dalam 90 hari terakhir.',
    'checkin_consistency',
    60,
    true,
    now(),
    now()
  ),
  (
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'journal_30_of_60',
    'Ritme Jurnal Stabil',
    'Tulis minimal 30 jurnal dalam 60 hari terakhir.',
    'journal_consistency',
    30,
    true,
    now(),
    now()
  ),
  (
    'abababab-abab-abab-abab-abababababab',
    'community_participation_25',
    'Kolaborator Komunitas',
    'Capai total 25 interaksi komunitas yang konstruktif.',
    'community_participation',
    25,
    true,
    now(),
    now()
  )
ON CONFLICT (code) DO NOTHING;
