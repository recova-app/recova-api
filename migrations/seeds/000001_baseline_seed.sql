BEGIN;

-- 1) Users
INSERT INTO users (
  id,
  google_id,
  email,
  username,
  password_hash,
  nickname,
  user_why,
  porn_free_goal,
  check_in_time,
  created_at,
  updated_at
)
VALUES
  ('10000000-0000-0000-0000-000000000001', 'google-oauth2|andre-seed-001', 'andre.wijaya@gmail.com', 'andre_wijaya', '$2a$10$gT8OqnXWNmeKUTUDJCYSN.T80XB41R3WT.IWl3UVYomJ6k5VOvbAe', 'Andre', 'Saya ingin pulih dari kecanduan ini untuk memperbaiki hubungan dengan pasangan dan menjadi pribadi yang lebih baik bagi keluarga saya.', 90, '07:00', now() - interval '120 days', now()),
  ('10000000-0000-0000-0000-000000000002', 'google-oauth2|budi-seed-002', 'budi.santoso@gmail.com', 'budi_santoso', '$2a$10$gT8OqnXWNmeKUTUDJCYSN.T80XB41R3WT.IWl3UVYomJ6k5VOvbAe', 'Budi', 'Ingin membangun hubungan yang sehat dan bermakna agar hidup lebih stabil dan bertanggung jawab.', 60, '06:30', now() - interval '118 days', now()),
  ('10000000-0000-0000-0000-000000000003', NULL, 'david.chen@yahoo.com', 'david_chen', '$2a$10$gT8OqnXWNmeKUTUDJCYSN.T80XB41R3WT.IWl3UVYomJ6k5VOvbAe', 'David', 'Saya ingin kembali produktif dan keluar dari pola yang selama ini menghambat karier dan kualitas hidup.', 120, '08:00', now() - interval '115 days', now()),
  ('10000000-0000-0000-0000-000000000004', NULL, 'ryan.pratama@gmail.com', 'ryan_pratama', '$2a$10$gT8OqnXWNmeKUTUDJCYSN.T80XB41R3WT.IWl3UVYomJ6k5VOvbAe', 'Ryan', 'Saya ingin kesehatan mental lebih stabil, emosi lebih terkendali, dan hubungan sosial lebih sehat.', 45, '06:00', now() - interval '110 days', now()),
  ('10000000-0000-0000-0000-000000000005', NULL, 'faisal.rahman@outlook.com', 'faisal_rahman', '$2a$10$gT8OqnXWNmeKUTUDJCYSN.T80XB41R3WT.IWl3UVYomJ6k5VOvbAe', 'Faisal', 'Saya ingin fokus akademik kembali kuat, menyelesaikan studi tepat waktu, dan membangun masa depan lebih baik.', 75, '07:30', now() - interval '108 days', now()),
  ('10000000-0000-0000-0000-000000000006', NULL, 'eric.tan@gmail.com', 'eric_tan', '$2a$10$gT8OqnXWNmeKUTUDJCYSN.T80XB41R3WT.IWl3UVYomJ6k5VOvbAe', 'Eric', 'Saya ingin hidup lebih selaras dengan nilai pribadi dan spiritual, dengan kebiasaan harian yang lebih sehat.', 100, '05:30', now() - interval '105 days', now())
ON CONFLICT (email) DO UPDATE
SET
  username = EXCLUDED.username,
  password_hash = EXCLUDED.password_hash,
  google_id = EXCLUDED.google_id,
  email = EXCLUDED.email,
  nickname = EXCLUDED.nickname,
  user_why = EXCLUDED.user_why,
  porn_free_goal = EXCLUDED.porn_free_goal,
  check_in_time = EXCLUDED.check_in_time,
  updated_at = now();

-- 2) Profiles
INSERT INTO profiles (
  id,
  user_id,
  answers,
  dependency_level,
  ai_summary,
  created_at,
  updated_at
)
VALUES
  ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '{"durasi_kecanduan":"5-6 tahun","frekuensi_harian":"2-3 kali per hari","trigger_utama":"Stres kerja, kesepian, media sosial","support_system":"Pasangan, konselor online, grup recovery"}'::jsonb, 'High', 'Andre menunjukkan motivasi tinggi dan konsistensi baik. Fokus lanjutan: manajemen trigger malam hari dan rutinitas pemulihan setelah kerja.', now() - interval '119 days', now()),
  ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', '{"durasi_kecanduan":"4-5 tahun","frekuensi_harian":"1-2 kali per hari","trigger_utama":"Waktu luang berlebih, scroll konten acak","support_system":"Teman dekat, komunitas online"}'::jsonb, 'High', 'Budi stabil dan progresif. Fokus lanjutan: sistem accountability harian dan evaluasi mingguan berbasis data check-in.', now() - interval '117 days', now()),
  ('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003', '{"durasi_kecanduan":"8-9 tahun","frekuensi_harian":"3-5 kali per hari","trigger_utama":"Prokrastinasi, stres performa, isolasi","support_system":"Psikolog, mentor, grup support"}'::jsonb, 'High', 'David berada di fase pemulihan menengah dengan beberapa relapse periodik. Fokus lanjutan: struktur kerja blok waktu dan recovery plan saat warning signal muncul.', now() - interval '114 days', now()),
  ('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000004', '{"durasi_kecanduan":"6-7 tahun","frekuensi_harian":"2-4 kali per hari","trigger_utama":"Kecemasan, insomnia, emosi negatif","support_system":"Psikiater, support group, keluarga"}'::jsonb, 'High', 'Ryan sudah menunjukkan penguatan coping skill. Fokus lanjutan: protokol malam hari, sleep hygiene, dan emergency routine saat urge tinggi.', now() - interval '109 days', now()),
  ('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000005', '{"durasi_kecanduan":"3-4 tahun","frekuensi_harian":"1-3 kali per hari","trigger_utama":"Deadline kuliah, tekanan nilai, distraksi digital","support_system":"Senior kampus, konselor, teman komunitas"}'::jsonb, 'Medium', 'Faisal punya motivasi akademik kuat. Fokus lanjutan: disiplin jadwal belajar, social support terstruktur, dan kontrol paparan konten pemicu.', now() - interval '107 days', now()),
  ('20000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000006', '{"durasi_kecanduan":"7-8 tahun","frekuensi_harian":"2-3 kali per hari","trigger_utama":"Rasa bersalah, kesepian, stres relasi","support_system":"Mentor spiritual, grup refleksi, sahabat"}'::jsonb, 'High', 'Eric konsisten pada pendekatan refleksi diri. Fokus lanjutan: mempertahankan journaling, menjaga ritme tidur, dan menurunkan self-blame berlebih.', now() - interval '104 days', now())
ON CONFLICT (user_id) DO UPDATE
SET
  answers = EXCLUDED.answers,
  dependency_level = EXCLUDED.dependency_level,
  ai_summary = EXCLUDED.ai_summary,
  updated_at = now();

-- 3) Streak history
INSERT INTO streaks (
  id,
  user_id,
  start_date,
  end_date,
  is_active,
  created_at,
  updated_at
)
VALUES
  ('30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', now() - interval '45 days', now() - interval '15 days', false, now() - interval '45 days', now()),
  ('30000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', now() - interval '14 days', NULL, true, now() - interval '14 days', now()),
  ('30000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', now() - interval '21 days', NULL, true, now() - interval '21 days', now()),
  ('30000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000003', now() - interval '60 days', now() - interval '22 days', false, now() - interval '60 days', now()),
  ('30000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000003', now() - interval '21 days', NULL, true, now() - interval '21 days', now()),
  ('30000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000004', now() - interval '90 days', now() - interval '75 days', false, now() - interval '90 days', now()),
  ('30000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000004', now() - interval '50 days', now() - interval '15 days', false, now() - interval '50 days', now()),
  ('30000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000004', now() - interval '14 days', NULL, true, now() - interval '14 days', now()),
  ('30000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000005', now() - interval '28 days', NULL, true, now() - interval '28 days', now()),
  ('30000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000006', now() - interval '35 days', now() - interval '8 days', false, now() - interval '35 days', now()),
  ('30000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000006', now() - interval '7 days', NULL, true, now() - interval '7 days', now())
ON CONFLICT (id) DO UPDATE
SET
  start_date = EXCLUDED.start_date,
  end_date = EXCLUDED.end_date,
  is_active = EXCLUDED.is_active,
  updated_at = now();

-- 4) Check-ins (14 hari x 6 user)
WITH checkin_seed (user_id, day_offset, mood, is_successful) AS (
  VALUES
    ('10000000-0000-0000-0000-000000000001', 0, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000001', 1, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000001', 2, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000001', 3, 'Sedikit Cemas', true),
    ('10000000-0000-0000-0000-000000000001', 4, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000001', 5, 'Lelah', false),
    ('10000000-0000-0000-0000-000000000001', 6, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000001', 7, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000001', 8, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000001', 9, 'Cemas', false),
    ('10000000-0000-0000-0000-000000000001', 10, 'Bingung', false),
    ('10000000-0000-0000-0000-000000000001', 11, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000001', 12, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000001', 13, 'Fokus', true),

    ('10000000-0000-0000-0000-000000000002', 0, 'Segar', true),
    ('10000000-0000-0000-0000-000000000002', 1, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000002', 2, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000002', 3, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000002', 4, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000002', 5, 'Segar', true),
    ('10000000-0000-0000-0000-000000000002', 6, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000002', 7, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000002', 8, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000002', 9, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000002', 10, 'Segar', true),
    ('10000000-0000-0000-0000-000000000002', 11, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000002', 12, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000002', 13, 'Fokus', true),

    ('10000000-0000-0000-0000-000000000003', 0, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000003', 1, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000003', 2, 'Cemas', false),
    ('10000000-0000-0000-0000-000000000003', 3, 'Sedih', false),
    ('10000000-0000-0000-0000-000000000003', 4, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000003', 5, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000003', 6, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000003', 7, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000003', 8, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000003', 9, 'Cemas', false),
    ('10000000-0000-0000-0000-000000000003', 10, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000003', 11, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000003', 12, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000003', 13, 'Tenang', true),

    ('10000000-0000-0000-0000-000000000004', 0, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000004', 1, 'Cemas', true),
    ('10000000-0000-0000-0000-000000000004', 2, 'Frustasi', false),
    ('10000000-0000-0000-0000-000000000004', 3, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000004', 4, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000004', 5, 'Gelisah', true),
    ('10000000-0000-0000-0000-000000000004', 6, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000004', 7, 'Cemas', false),
    ('10000000-0000-0000-0000-000000000004', 8, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000004', 9, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000004', 10, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000004', 11, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000004', 12, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000004', 13, 'Termotivasi', true),

    ('10000000-0000-0000-0000-000000000005', 0, 'Segar', true),
    ('10000000-0000-0000-0000-000000000005', 1, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000005', 2, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000005', 3, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000005', 4, 'Lelah', true),
    ('10000000-0000-0000-0000-000000000005', 5, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000005', 6, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000005', 7, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000005', 8, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000005', 9, 'Segar', true),
    ('10000000-0000-0000-0000-000000000005', 10, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000005', 11, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000005', 12, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000005', 13, 'Fokus', true),

    ('10000000-0000-0000-0000-000000000006', 0, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000006', 1, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000006', 2, 'Gelisah', false),
    ('10000000-0000-0000-0000-000000000006', 3, 'Frustasi', false),
    ('10000000-0000-0000-0000-000000000006', 4, 'Cemas', false),
    ('10000000-0000-0000-0000-000000000006', 5, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000006', 6, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000006', 7, 'Bahagia', true),
    ('10000000-0000-0000-0000-000000000006', 8, 'Tenang', true),
    ('10000000-0000-0000-0000-000000000006', 9, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000006', 10, 'Gelisah', false),
    ('10000000-0000-0000-0000-000000000006', 11, 'Fokus', true),
    ('10000000-0000-0000-0000-000000000006', 12, 'Termotivasi', true),
    ('10000000-0000-0000-0000-000000000006', 13, 'Bahagia', true)
)
INSERT INTO check_ins (
  id,
  user_id,
  check_in_date,
  mood,
  is_successful,
  relapse_trigger,
  created_at
)
SELECT
  gen_random_uuid(),
  cs.user_id::uuid,
  current_date - cs.day_offset,
  cs.mood,
  cs.is_successful,
  CASE
    WHEN cs.is_successful THEN NULL
    WHEN cs.mood IN ('Cemas', 'Sedikit Cemas', 'Gelisah') THEN ARRAY['stres kerja', 'kecemasan malam']
    WHEN cs.mood IN ('Frustasi', 'Sedih', 'Bingung') THEN ARRAY['emosi negatif', 'isolasi']
    ELSE ARRAY['kelelahan', 'paparan media sosial']
  END::text[],
  now() - make_interval(days => cs.day_offset)
FROM checkin_seed cs
ON CONFLICT (user_id, check_in_date) DO UPDATE
SET
  mood = EXCLUDED.mood,
  is_successful = EXCLUDED.is_successful,
  relapse_trigger = EXCLUDED.relapse_trigger;

-- 5) Journals (1:1 dengan check-in)
WITH journal_source AS (
  SELECT
    ci.user_id,
    ci.id AS check_in_id,
    ci.check_in_date,
    ci.mood,
    ci.is_successful
  FROM check_ins ci
  WHERE ci.user_id IN (
    '10000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000003',
    '10000000-0000-0000-0000-000000000004',
    '10000000-0000-0000-0000-000000000005',
    '10000000-0000-0000-0000-000000000006'
  )
    AND ci.check_in_date BETWEEN current_date - 13 AND current_date
)
INSERT INTO journals (
  id,
  user_id,
  check_in_id,
  content,
  created_at
)
SELECT
  gen_random_uuid(),
  js.user_id,
  js.check_in_id,
  CASE
    WHEN js.is_successful THEN
      format(
        'Refleksi %s: hari ini berhasil dengan mood %s. Fokus berikutnya: tidur cukup, olahraga ringan, dan check-in tepat waktu.',
        to_char(js.check_in_date, 'YYYY-MM-DD'),
        js.mood
      )
    ELSE
      format(
        'Refleksi %s: hari ini belum berhasil (mood %s). Rencana pemulihan: evaluasi pemicu, hubungi support system, dan reset rencana besok.',
        to_char(js.check_in_date, 'YYYY-MM-DD'),
        js.mood
      )
  END,
  (js.check_in_date::timestamp + interval '21 hours')
FROM journal_source js
ON CONFLICT (check_in_id) DO UPDATE
SET
  content = EXCLUDED.content;

-- 6) Community posts
INSERT INTO community_posts (
  id,
  user_id,
  title,
  content,
  category,
  comment_count,
  like_count,
  created_at,
  updated_at
)
VALUES
  ('50000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'Tips Mengatasi Urge di Minggu Pertama', 'Saya berbagi pola yang paling membantu di minggu pertama: aktivitas fisik segera, pindah ke ruang publik, jeda napas, dan batasi akses konten pemicu.', 'saran', 0, 0, now() - interval '20 days', now()),
  ('50000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000004', 'Hari ke-21 dan Brain Fog Masih Kuat', 'Sudah 3 minggu clean tetapi brain fog masih terasa. Ada yang pernah di fase ini? Butuh insight cara bertahan tanpa panik berlebihan.', 'pertanyaan', 0, 0, now() - interval '19 days', now()),
  ('50000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000005', 'Sukses 30 Hari Konsisten', 'Saya berhasil 30 hari konsisten. Tidur membaik, fokus meningkat, dan emosi lebih stabil.', 'motivasi', 0, 0, now() - interval '18 days', now()),
  ('50000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000006', 'Relapse Setelah 14 Hari, Cara Bangkit Lagi?', 'Kemarin saya relapse setelah 14 hari. Sekarang sedang reset lagi. Mohon saran konkret untuk memutus pola yang sama saat malam hari.', 'bantuan', 0, 0, now() - interval '17 days', now()),
  ('50000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', 'Mencari Accountability Partner Harian', 'Saya ingin bentuk grup kecil untuk saling check-in harian. Fokusnya saling dukung, bukan saling menghakimi.', 'bantuan', 0, 0, now() - interval '16 days', now()),
  ('50000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000003', 'Strategi Menghadapi Trigger dari Media Sosial', 'Saya sudah unfollow banyak akun pemicu, tapi explore masih sering memunculkan konten berisiko. Praktik apa paling efektif?', 'pertanyaan', 0, 0, now() - interval '15 days', now()),
  ('50000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002', 'Perubahan Positif Setelah 60 Hari', 'Setelah 60 hari, saya merasa energi lebih stabil, relasi membaik, dan produktivitas meningkat.', 'cerita', 0, 0, now() - interval '14 days', now()),
  ('50000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', 'Rekomendasi Konten Edukatif Recovery', 'Teman-teman, boleh share buku, video, atau podcast yang membantu membangun mindset recovery?', 'saran', 0, 0, now() - interval '13 days', now()),
  ('50000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000003', 'Template Recovery Plan Mingguan', 'Saya bagikan template sederhana: target check-in, target tidur, target olahraga, dan trigger log.', 'saran', 0, 0, now() - interval '12 days', now()),
  ('50000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000004', 'Cerita: Dari Isolasi ke Komunikasi Terbuka', 'Dulu saya menutup diri. Setelah jujur ke support system, tekanan berkurang signifikan.', 'cerita', 0, 0, now() - interval '11 days', now()),
  ('50000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000005', 'Butuh Saran Saat Deadline Menumpuk', 'Saat deadline kuliah menumpuk, urge meningkat. Bagaimana menjaga fokus belajar tanpa burnout?', 'bantuan', 0, 0, now() - interval '10 days', now()),
  ('50000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000006', 'Pertanyaan: Menjaga Konsistensi Setelah Relapse', 'Apa indikator paling penting untuk memastikan kita benar-benar pulih setelah relapse?', 'pertanyaan', 0, 0, now() - interval '9 days', now()),
  ('50000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000001', 'Update 45 Hari: Lebih Stabil Secara Emosi', 'Masuk hari ke-45, emosi mulai lebih stabil dan pola tidur membaik. Yang paling membantu: kurangi scroll malam dan journaling singkat.', 'cerita', 0, 0, now() - interval '8 days', now()),
  ('50000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000002', 'Tips Saat Urge Datang Tiba-Tiba', 'Saya pakai pola 3 langkah: tarik napas, pindah lokasi, lalu mulai aktivitas fisik ringan 10 menit. Cukup efektif buat meredakan dorongan awal.', 'saran', 0, 0, now() - interval '8 days' + interval '3 hours', now()),
  ('50000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000003', 'Butuh Saran Menghadapi Weekend', 'Di akhir pekan urge sering naik karena waktu kosong. Kalian biasanya bikin struktur aktivitas seperti apa?', 'bantuan', 0, 0, now() - interval '7 days' + interval '2 hours', now()),
  ('50000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000004', 'Relapse Log: Belajar dari Pola yang Sama', 'Saya catat relapse terakhir terjadi saat begadang + stres kerja. Sekarang lagi coba fokus ke sleep hygiene dan batas layar malam.', 'cerita', 0, 0, now() - interval '7 days' + interval '6 hours', now()),
  ('50000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000005', 'Mencari Rekomendasi Podcast Recovery', 'Ada rekomendasi podcast atau channel yang bahas recovery PMO secara praktis dan tidak menghakimi?', 'pertanyaan', 0, 0, now() - interval '6 days' + interval '1 hours', now()),
  ('50000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000006', 'Checklist Harian yang Membantu Saya', 'Checklist sederhana: tidur cukup, olahraga 20 menit, no-scroll malam, dan check-in emosi. Kecil, tapi dampaknya terasa.', 'saran', 0, 0, now() - interval '6 days' + interval '4 hours', now()),
  ('50000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000002', 'Hari Sulit Bukan Berarti Gagal', 'Kemarin hari berat dan hampir relapse, tapi berhasil stop sebelum terlambat. Mau ngingetin: progres itu juga tentang cara kita bangkit.', 'motivasi', 0, 0, now() - interval '5 days' + interval '2 hours', now()),
  ('50000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000003', 'Template No-Social Window Malam', 'Saya coba aturan tanpa media sosial setelah jam 20.30. Hasilnya pikiran lebih tenang dan tidur lebih cepat.', 'saran', 0, 0, now() - interval '4 days' + interval '1 hours', now()),
  ('50000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000004', 'Pertanyaan: Menjaga Fokus Saat WFH', 'Saat kerja dari rumah, trigger digital lebih sering muncul. Kalian punya setup kerja yang efektif untuk meminimalkan distraksi?', 'pertanyaan', 0, 0, now() - interval '3 days' + interval '3 hours', now()),
  ('50000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000001', 'Ajak Bikin Grup Check-in Pagi', 'Kalau ada yang mau, kita bikin grup kecil check-in pagi 5 menit biar saling jaga ritme sebelum mulai aktivitas.', 'bantuan', 0, 0, now() - interval '2 days' + interval '5 hours', now())
ON CONFLICT (id) DO UPDATE
SET
  user_id = EXCLUDED.user_id,
  title = EXCLUDED.title,
  content = EXCLUDED.content,
  category = EXCLUDED.category,
  updated_at = now();

-- 7) Community comments + replies
INSERT INTO community_comments (
  id,
  user_id,
  post_id,
  parent_comment_id,
  content,
  depth,
  reply_count,
  created_at
)
VALUES
  ('60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000001', NULL, 'Terima kasih, tips ini kepakai banget terutama saat sore hari.', 0, 0, now() - interval '20 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000001', NULL, 'Setuju, aktivitas fisik cepat memang paling efektif untuk memutus dorongan awal.', 0, 0, now() - interval '20 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000001', NULL, 'Saya tambahkan teknik napas 4-7-8, membantu menurunkan impuls dalam 2-3 menit.', 0, 0, now() - interval '20 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000002', NULL, 'Normal di fase awal rewiring. Fokus ke tidur, hidrasi, dan konsistensi check-in.', 0, 0, now() - interval '19 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000002', NULL, 'Saya dulu baru membaik signifikan sekitar minggu ke-6. Tetap lanjutkan proses.', 0, 0, now() - interval '19 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000003', NULL, 'Selamat! Konsistensi kecil harian memang bikin dampak besar.', 0, 0, now() - interval '18 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000003', NULL, 'Boleh share tiga kebiasaan utama yang paling membantu?', 0, 0, now() - interval '18 days' + interval '5 hours'),
  ('60000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000004', NULL, 'Relapse bukan akhir. Yang penting sekarang catat trigger dan rancang respon baru.', 0, 0, now() - interval '17 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000004', NULL, 'Coba siapkan emergency routine 15 menit yang selalu sama setiap urge naik.', 0, 0, now() - interval '17 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000005', NULL, 'Saya ikut. Bisa mulai dengan check-in jam tetap dan evaluasi mingguan singkat.', 0, 0, now() - interval '16 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000005', NULL, 'Saya juga berminat, terutama untuk accountability soal jam tidur dan fokus belajar.', 0, 0, now() - interval '16 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000006', NULL, 'Saya pakai kombinasi limit aplikasi, unfollow ketat, dan blokir kata kunci.', 0, 0, now() - interval '15 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000006', NULL, 'Tambahkan no-social window setelah jam 20.00, dampaknya signifikan.', 0, 0, now() - interval '15 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000007', NULL, 'Cerita ini sangat menguatkan. Terima kasih sudah berbagi progres real.', 0, 0, now() - interval '14 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000007', NULL, 'Saya baru hari ke-12 dan mulai merasakan energi membaik. Semoga konsisten.', 0, 0, now() - interval '14 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000008', NULL, 'Untuk buku, saya sarankan fokus pada habit, emotional regulation, dan deep work.', 0, 0, now() - interval '13 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000008', NULL, 'Podcast yang praktis membantu saat commute, terutama bahasan self-regulation.', 0, 0, now() - interval '13 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000009', NULL, 'Template ini bagus. Saya sudah coba dan review mingguan jadi lebih objektif.', 0, 0, now() - interval '12 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000010', NULL, 'Komunikasi terbuka memang berat di awal, tapi dampaknya sangat besar ke kestabilan emosi.', 0, 0, now() - interval '11 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000011', NULL, 'Saat deadline tinggi, saya pecah tugas ke blok 45 menit + jeda aktif 10 menit.', 0, 0, now() - interval '10 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000012', NULL, 'Indikator utama: recovery plan tetap jalan 7 hari setelah relapse, bukan nol total.', 0, 0, now() - interval '9 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', 'Setuju, saya pakai pola serupa dengan tambahan journaling 5 menit tiap malam.', 1, 0, now() - interval '19 days' + interval '5 hours'),
  ('60000000-0000-0000-0000-000000000023', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000022', 'Mantap, journaling pendek memang sangat membantu buat awareness emosi.', 2, 0, now() - interval '19 days' + interval '6 hours'),
  ('60000000-0000-0000-0000-000000000024', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000008', 'Kalau perlu, buat daftar orang yang bisa dihubungi dalam 10 menit saat krisis.', 1, 0, now() - interval '16 days' + interval '6 hours'),
  ('60000000-0000-0000-0000-000000000025', '10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000011', '60000000-0000-0000-0000-000000000020', 'Saya tambahkan aturan no-scroll sebelum deadline agar fokus tetap aman.', 1, 0, now() - interval '9 days' + interval '5 hours'),
  ('60000000-0000-0000-0000-000000000026', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000013', NULL, 'Keren progresnya. Saya juga ngerasa tidur lebih rapi bikin urge jauh lebih rendah.', 0, 0, now() - interval '8 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000027', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000013', NULL, 'Setuju soal journaling pendek, 5 menit aja sudah bantu menenangkan pikiran.', 0, 0, now() - interval '8 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000028', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000014', NULL, 'Teknik 3 langkah ini praktis banget. Saya biasa tambah minum air dulu sebelum pindah lokasi.', 0, 0, now() - interval '8 days' + interval '5 hours'),
  ('60000000-0000-0000-0000-000000000029', '10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000014', '60000000-0000-0000-0000-000000000028', 'Nice tambahan. Saya juga pakai timer 10 menit biar tetap konsisten.', 1, 0, now() - interval '8 days' + interval '6 hours'),
  ('60000000-0000-0000-0000-000000000030', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000015', NULL, 'Weekend saya isi blok waktu: olahraga pagi, ketemu teman, dan review mingguan malam.', 0, 0, now() - interval '7 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000031', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000015', NULL, 'Coba bikin rencana dari Jumat malam supaya Sabtu-Minggu tidak kosong.', 0, 0, now() - interval '7 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000032', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000016', NULL, 'Bagus banget kamu sudah catat polanya. Sleep hygiene biasanya jadi game changer.', 0, 0, now() - interval '7 days' + interval '7 hours'),
  ('60000000-0000-0000-0000-000000000033', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000016', '60000000-0000-0000-0000-000000000032', 'Setuju. Saya bantu dengan alarm winding-down 1 jam sebelum tidur.', 1, 0, now() - interval '7 days' + interval '8 hours'),
  ('60000000-0000-0000-0000-000000000034', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000017', NULL, 'Podcast yang menekankan habit dan regulasi emosi paling membantu buat saya.', 0, 0, now() - interval '6 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000035', '10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000018', NULL, 'Checklist harian ini simpel dan realistis. Saya mau adaptasi untuk jam kerja saya.', 0, 0, now() - interval '6 days' + interval '5 hours'),
  ('60000000-0000-0000-0000-000000000036', '10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000018', '60000000-0000-0000-0000-000000000035', 'Sip, saya tambahkan kolom mood biar bisa lihat pola emosinya juga.', 1, 0, now() - interval '6 days' + interval '6 hours'),
  ('60000000-0000-0000-0000-000000000037', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000019', NULL, 'Respect, bisa stop sebelum relapse itu kemenangan besar. Keep going.', 0, 0, now() - interval '5 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000038', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000019', NULL, 'Setuju, yang penting bukan sempurna, tapi cepat kembali ke rencana.', 0, 0, now() - interval '5 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000039', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000020', NULL, 'No-social window malam memang powerful. Saya pakai dari 20.00 sampai tidur.', 0, 0, now() - interval '4 days' + interval '2 hours'),
  ('60000000-0000-0000-0000-000000000040', '10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000020', '60000000-0000-0000-0000-000000000039', 'Sama, saya kombinasi dengan mode grayscale supaya less tempting.', 1, 0, now() - interval '4 days' + interval '3 hours'),
  ('60000000-0000-0000-0000-000000000041', '10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000021', NULL, 'Untuk WFH, saya pisahkan akun kerja dan akun personal di browser.', 0, 0, now() - interval '3 days' + interval '4 hours'),
  ('60000000-0000-0000-0000-000000000042', '10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000021', NULL, 'Coba juga pakai website blocker saat jam fokus utama.', 0, 0, now() - interval '3 days' + interval '5 hours'),
  ('60000000-0000-0000-0000-000000000043', '10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000022', NULL, 'Saya ikut grup check-in pagi. Format 5 menit cocok buat jaga konsistensi.', 0, 0, now() - interval '2 days' + interval '6 hours'),
  ('60000000-0000-0000-0000-000000000044', '10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000022', NULL, 'Ikut juga. Mungkin bisa pakai template: mood, target, dan 1 komitmen hari ini.', 0, 0, now() - interval '2 days' + interval '7 hours'),
  ('60000000-0000-0000-0000-000000000045', '10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000022', '60000000-0000-0000-0000-000000000044', 'Setuju, besok saya mulai dengan format itu ya.', 1, 0, now() - interval '2 days' + interval '8 hours')
ON CONFLICT (id) DO UPDATE
SET
  user_id = EXCLUDED.user_id,
  post_id = EXCLUDED.post_id,
  parent_comment_id = EXCLUDED.parent_comment_id,
  content = EXCLUDED.content,
  depth = EXCLUDED.depth,
  created_at = EXCLUDED.created_at;

-- 8) Community likes
INSERT INTO community_post_likes (user_id, post_id, created_at)
VALUES
  ('10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000001', now() - interval '20 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000001', now() - interval '20 days' + interval '7 hours'),
  ('10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000001', now() - interval '20 days' + interval '8 hours'),
  ('10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000002', now() - interval '19 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000002', now() - interval '19 days' + interval '7 hours'),
  ('10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000003', now() - interval '18 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000003', now() - interval '18 days' + interval '7 hours'),
  ('10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000004', now() - interval '17 days' + interval '5 hours'),
  ('10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000004', now() - interval '17 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000005', now() - interval '16 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000005', now() - interval '16 days' + interval '7 hours'),
  ('10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000006', now() - interval '15 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000007', now() - interval '14 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000007', now() - interval '14 days' + interval '7 hours'),
  ('10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000008', now() - interval '13 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000004', '50000000-0000-0000-0000-000000000008', now() - interval '13 days' + interval '7 hours'),
  ('10000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000009', now() - interval '12 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000010', now() - interval '11 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000011', now() - interval '10 days' + interval '6 hours'),
  ('10000000-0000-0000-0000-000000000003', '50000000-0000-0000-0000-000000000012', now() - interval '9 days' + interval '6 hours')
ON CONFLICT (user_id, post_id) DO NOTHING;

-- 9) Recompute derived counters
UPDATE community_comments parent
SET reply_count = child.total_reply
FROM (
  SELECT
    parent_comment_id,
    COUNT(*)::int AS total_reply
  FROM community_comments
  WHERE parent_comment_id IS NOT NULL
  GROUP BY parent_comment_id
) child
WHERE parent.id = child.parent_comment_id;

UPDATE community_comments
SET reply_count = 0
WHERE parent_comment_id IS NULL
  AND id NOT IN (
    SELECT DISTINCT parent_comment_id
    FROM community_comments
    WHERE parent_comment_id IS NOT NULL
  );

UPDATE community_posts p
SET
  comment_count = counts.comment_count,
  like_count = counts.like_count,
  updated_at = now()
FROM (
  SELECT
    cp.id,
    COALESCE(comment_totals.comment_count, 0) AS comment_count,
    COALESCE(like_totals.like_count, 0) AS like_count
  FROM community_posts cp
  LEFT JOIN (
    SELECT post_id, COUNT(*)::int AS comment_count
    FROM community_comments
    GROUP BY post_id
  ) comment_totals ON comment_totals.post_id = cp.id
  LEFT JOIN (
    SELECT post_id, COUNT(*)::int AS like_count
    FROM community_post_likes
    GROUP BY post_id
  ) like_totals ON like_totals.post_id = cp.id
) counts
WHERE p.id = counts.id;

-- 10) Education contents
INSERT INTO education_contents (
  id,
  title,
  description,
  url,
  thumbnail_url,
  category,
  type,
  is_active,
  published_at
)
VALUES
  ('18181818-1818-1818-1818-181818181818', 'Kecanduan Pornografi: Penyebab, Gejala, dan Perawatan', 'Membahas tanda kecanduan, perubahan kerja otak, serta opsi penanganan medis dan psikologis.', 'https://www.alodokter.com/kecanduan-pornografi-penyebab-gejala-dan-perawatan', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '50 days'),
  ('19191919-1919-1919-1919-191919191919', '6 Dampak Menonton Video Porno yang Tidak Boleh Disepelekan', 'Merangkum dampak pada tanggung jawab, relasi, produktivitas, dan kesehatan mental akibat kebiasaan menonton berlebihan.', 'https://www.alodokter.com/dampak-buruk-yang-dapat-dialami-penggemar-video-porno', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '49 days'),
  ('20202020-2020-2020-2020-202020202020', 'Aktivitas Otak Pecandu Pornografi Mirip Pecandu Narkoba', 'Ulasan populer tentang respons sistem reward otak pada konsumsi pornografi berulang.', 'https://tirto.id/aktivitas-otak-pecandu-pornografi-mirip-pecandu-narkoba-ch4R', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '48 days'),
  ('21212121-2121-2121-2121-212121212121', 'Nonton Film Porno Itu Wajar, Tapi Waspadai Juga Segudang Risikonya', 'Menjelaskan batas wajar konsumsi, ciri kecanduan, dan risiko bila kebiasaan menjadi berlebihan.', 'https://hellosehat.com/mental/kecanduan/pria-suka-nonton-film-porno/', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '47 days'),
  ('23232323-2323-2323-2323-232323232323', 'Tanda-Tanda Kecanduan Pornografi dan Cara Mengatasinya', 'Panduan mengenali gejala adiksi serta langkah pemulihan yang bisa diterapkan bertahap.', 'https://hellosehat.com/mental/kecanduan/kecanduan-pornografi/', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '46 days'),
  ('24242424-2424-2424-2424-242424242424', 'Disfungsi Ereksi Akibat Kecanduan Film Porno, Mungkinkah?', 'Membahas kaitan paparan konten dewasa berlebih dengan risiko gangguan ereksi dan fungsi seksual.', 'https://www.klikdokter.com/gaya-hidup/perawatan-pria/disfungsi-ereksi-akibat-kecanduan-film-porno-mungkinkah', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '45 days'),
  ('25252525-2525-2525-2525-252525252525', 'Awas, Pria yang Kecanduan Pornografi Malah Berisiko Alami Impotensi', 'Menjelaskan mekanisme penurunan sensitivitas rangsangan seksual akibat konsumsi porno berlebihan.', 'https://hellosehat.com/pria/penyakit-pria/kecanduan-pornografi-risiko-impoten/', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '44 days'),
  ('26262626-2626-2626-2626-262626262626', 'Terapi untuk Remaja yang Kecanduan Pornografi', 'Membahas opsi terapi untuk remaja, termasuk modifikasi perilaku, psikoterapi, dan pendampingan keluarga.', 'https://www.klikdokter.com/psikologi/kesehatan-mental/terapi-untuk-remaja-yang-kecanduan-pornografi', NULL, 'Dampak Pornografi', 'artikel', true, now() - interval '43 days'),
  ('27272727-2727-2727-2727-272727272727', 'Strategi Mencegah dan Mengatasi Kecanduan Porno', 'Membahas langkah pencegahan, CBT, dukungan sosial, dan rutinitas sehat untuk pemulihan.', 'https://www.klikdokter.com/psikologi/kesehatan-mental/menavigasi-lautan-digital-strategi-menarik-untuk-mencegah-dan-mengatasi-kecanduan-porno', NULL, 'Pemulihan', 'artikel', true, now() - interval '42 days'),
  ('28282828-2828-2828-2828-282828282828', 'Ini Efek Kecanduan Pornografi pada Kesehatan Mental', 'Membahas dampak ke kecemasan, mood, produktivitas, dan kualitas relasi sehari-hari.', 'https://www.halodoc.com/artikel/ini-efek-kecanduan-pornografi-pada-kesehatan-mental', NULL, 'Pemulihan', 'artikel', true, now() - interval '41 days'),
  ('29292929-2929-2929-2929-292929292929', 'Efek Kecanduan Pornografi pada Disfungsi Ereksi', 'Menjelaskan kaitan perubahan respons otak akibat konten pornografi dengan disfungsi ereksi.', 'https://www.halodoc.com/artikel/efek-kecanduan-pornografi-pada-disfungsi-ereksi', NULL, 'Pemulihan', 'artikel', true, now() - interval '40 days'),
  ('30303030-3030-3030-3030-303030303030', 'Dopamine Detox, Pilihan Metode untuk Mengatasi Kecanduan', 'Menjelaskan konsep dopamine detox sebagai strategi pembatasan stimulasi instan secara bertahap.', 'https://www.alodokter.com/dopamine-detox-pilihan-metode-untuk-mengatasi-kecanduan', NULL, 'Pemulihan', 'artikel', true, now() - interval '39 days'),
  ('31313131-3131-3131-3131-313131313131', 'Mengenal Dopamine Detox, Manfaat dan Cara Melakukannya', 'Membahas manfaat, batasan ilmiah, dan langkah praktis menerapkan dopamine detox dengan aman.', 'https://www.halodoc.com/artikel/mengenal-dopamine-detox-manfaat-dan-cara-melakukannya', NULL, 'Pemulihan', 'artikel', true, now() - interval '38 days'),
  ('32323232-3232-3232-3232-323232323232', 'Cara Jitu Agar Tidak Kecanduan PMO', 'Panduan langkah preventif dan intervensi dini agar siklus pornografi-masturbasi-orgasme tidak semakin berat.', 'https://www.halodoc.com/artikel/cara-jitu-agar-tidak-kecanduan-pmo', NULL, 'Pemulihan', 'artikel', true, now() - interval '37 days'),
  ('33333333-3333-3333-3333-333333333333', '7 Cara Berhenti Onani supaya Lepas dari Kecanduan', 'Strategi mengenali pemicu, menghindari konten pornografi, dan mencari bantuan profesional bila perlu.', 'https://www.alodokter.com/7-cara-berhenti-onani-supaya-lepas-dari-kecanduan', NULL, 'Pemulihan', 'artikel', true, now() - interval '36 days'),
  ('34343434-3434-3434-3434-343434343434', 'Pasangan Kecanduan Pornografi? Ini Cara Mengatasinya', 'Tips komunikasi, dukungan, dan batasan sehat untuk membantu pasangan keluar dari pola adiktif.', 'https://www.alodokter.com/pasangan-kecanduan-pornografi-ini-cara-mengatasinya', NULL, 'Pemulihan', 'artikel', true, now() - interval '35 days'),
  ('71000000-0000-0000-0000-000000000002', 'Langkah Pertama Berhenti dari Kecanduan Pornografi', 'Membahas pergeseran pola pikir awal dari penyangkalan ke tindakan pemulihan.', 'http://www.youtube.com/watch?v=a37iuykI9Io', 'https://i.ytimg.com/vi/a37iuykI9Io/maxresdefault.jpg', 'Kesehatan Mental', 'video', true, now() - interval '49 days'),
  ('71000000-0000-0000-0000-000000000003', 'Apa yang Terjadi Saat Berhenti Konsumsi Pornografi', 'Penjelasan timeline pemulihan fisik-mental setelah berhenti paparan konten adiktif.', 'http://www.youtube.com/watch?v=gJjsm2xcOy8', 'https://i.ytimg.com/vi/gJjsm2xcOy8/maxresdefault.jpg', 'Kesehatan Holistik', 'video', true, now() - interval '48 days'),
  ('71000000-0000-0000-0000-000000000004', 'Cara Memanage Nafsu - Ustadz Adi Hidayat', 'Kajian pengelolaan dorongan dari perspektif spiritual dan disiplin diri.', 'http://www.youtube.com/watch?v=TqZIsmrQ06o', 'https://i.ytimg.com/vi/TqZIsmrQ06o/maxresdefault.jpg', 'Spiritualitas', 'video', true, now() - interval '47 days'),
  ('71000000-0000-0000-0000-000000000005', 'Apa yang Terjadi Saat Orang Ketagihan Pornografi?', 'Penjelasan ilmiah dampak adiksi terhadap otak dan perilaku.', 'http://www.youtube.com/watch?v=Sq1s564ukTI', 'https://i.ytimg.com/vi/Sq1s564ukTI/maxresdefault.jpg', 'Kesehatan Mental', 'video', true, now() - interval '46 days'),
  ('71000000-0000-0000-0000-000000000006', 'Tips Disiplin Membangun Kebiasaan', 'Ringkasan prinsip habit design agar perubahan perilaku lebih sustain.', 'http://www.youtube.com/watch?v=uqGf4PWDOUw', 'https://i.ytimg.com/vi/uqGf4PWDOUw/maxresdefault.jpg', 'Pengembangan Diri', 'video', true, now() - interval '45 days'),
  ('71000000-0000-0000-0000-000000000007', 'Cara Reset Hidup dalam 7 Hari', 'Strategi reset rutinitas untuk memutus pola destruktif dan bangun arah baru.', 'http://www.youtube.com/watch?v=gPdKGv9ZuAU', 'https://i.ytimg.com/vi/gPdKGv9ZuAU/maxresdefault.jpg', 'Pengembangan Diri', 'video', true, now() - interval '44 days'),
  ('71000000-0000-0000-0000-000000000008', 'Rahasia Mengatasi Malas dan Kembali Produktif', 'Teknik praktis untuk memulai aksi kecil saat motivasi turun.', 'http://www.youtube.com/watch?v=WMfRHf5kjsE', 'https://i.ytimg.com/vi/WMfRHf5kjsE/maxresdefault.jpg', 'Produktivitas', 'video', true, now() - interval '43 days'),
  ('71000000-0000-0000-0000-000000000009', 'How to Be So Productive it Feels Illegal', 'Kumpulan habit produktivitas yang bisa diadaptasi ke rutinitas recovery.', 'http://www.youtube.com/watch?v=hSGt_rhu49U', 'https://i.ytimg.com/vi/hSGt_rhu49U/maxresdefault.jpg', 'Produktivitas', 'video', true, now() - interval '42 days'),
  ('71000000-0000-0000-0000-000000000010', 'Kunci Kebahagiaan dan Fokus Prioritas Hidup', 'Membahas pemilihan prioritas agar energi mental tidak habis pada hal tidak penting.', 'http://www.youtube.com/watch?v=dAI12OGD04A', 'https://i.ytimg.com/vi/dAI12OGD04A/maxresdefault.jpg', 'Kesadaran Diri', 'video', true, now() - interval '41 days'),
  ('71000000-0000-0000-0000-000000000011', 'Supaya Hidup Tidak Overthinking', 'Prinsip stoikisme praktis untuk mengelola pikiran berlebih.', 'http://www.youtube.com/watch?v=9qwR3GmR63I', 'https://i.ytimg.com/vi/9qwR3GmR63I/maxresdefault.jpg', 'Kesadaran Diri', 'video', true, now() - interval '40 days'),
  ('71000000-0000-0000-0000-000000000012', 'Cara Meningkatkan Fokus dan Kecerdasan Belajar', 'Strategi belajar terstruktur untuk memperkuat daya pikir.', 'http://www.youtube.com/watch?v=H-DeO-hnyTc', 'https://i.ytimg.com/vi/H-DeO-hnyTc/maxresdefault.jpg', 'Pengembangan Diri', 'video', true, now() - interval '39 days'),
  ('71000000-0000-0000-0000-000000000013', 'Mengatasi Rasa Kesepian', 'Panduan membangun koneksi sosial sehat untuk menurunkan risiko relapse.', 'http://www.youtube.com/watch?v=0b9Qzow_lv0', 'https://i.ytimg.com/vi/0b9Qzow_lv0/maxresdefault.jpg', 'Kesehatan Mental', 'video', true, now() - interval '38 days'),
  ('71000000-0000-0000-0000-000000000014', 'Yang Capek Jadi Dewasa, Nonton Ini', 'Membahas tekanan hidup dewasa dan langkah menjaga kestabilan emosi.', 'http://www.youtube.com/watch?v=ZOQhVk_YuSY', 'https://i.ytimg.com/vi/ZOQhVk_YuSY/maxresdefault.jpg', 'Kesehatan Mental', 'video', true, now() - interval '37 days'),
  ('71000000-0000-0000-0000-000000000015', 'Rahasia Jadi Manusia Bernilai', 'Insight pengembangan karakter dan keberanian mengambil tanggung jawab personal.', 'http://www.youtube.com/watch?v=E14rVsVJk0M', 'https://i.ytimg.com/vi/E14rVsVJk0M/maxresdefault.jpg', 'Pengembangan Diri', 'video', true, now() - interval '36 days')
ON CONFLICT (id) DO UPDATE
SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  url = EXCLUDED.url,
  thumbnail_url = EXCLUDED.thumbnail_url,
  category = EXCLUDED.category,
  type = EXCLUDED.type,
  is_active = EXCLUDED.is_active,
  published_at = EXCLUDED.published_at;

-- 11) Daily motivations
INSERT INTO daily_motivations (id, content, is_active, created_at)
VALUES
  ('30000000-0000-0000-0000-000000000001', 'Setiap hari adalah kesempatan baru untuk menjadi versi yang lebih baik dari dirimu.', true, now()),
  ('30000000-0000-0000-0000-000000000002', 'Kamu lebih kuat dari yang kamu kira. Setiap kali menolak godaan, kamu sedang membangun karakter.', true, now()),
  ('30000000-0000-0000-0000-000000000003', 'Relapse bukan akhir, yang penting adalah bangkit dan belajar dari pola yang terjadi.', true, now()),
  ('30000000-0000-0000-0000-000000000004', 'Perubahan dimulai dari keputusan kecil yang diulang setiap hari.', true, now()),
  ('30000000-0000-0000-0000-000000000005', 'Kamu tidak sendirian dalam perjuangan ini; dukungan selalu bisa dicari.', true, now()),
  ('30000000-0000-0000-0000-000000000006', 'Fokus pada hari ini. Satu hari konsisten adalah kemenangan nyata.', true, now()),
  ('30000000-0000-0000-0000-000000000007', 'Kesuksesan recovery diukur dari seberapa sering kamu bangkit setelah jatuh.', true, now()),
  ('30000000-0000-0000-0000-000000000008', 'Otakmu sedang beradaptasi. Bersabarlah pada proses penyembuhan.', true, now()),
  ('30000000-0000-0000-0000-000000000009', 'Setiap keputusan sehat hari ini memperkuat dirimu di masa depan.', true, now()),
  ('30000000-0000-0000-0000-000000000010', 'Kamu layak memiliki pikiran yang jernih dan hubungan yang sehat.', true, now()),
  ('30000000-0000-0000-0000-000000000011', 'Progress kadang tidak terlihat, tetapi prosesnya tetap berjalan.', true, now()),
  ('30000000-0000-0000-0000-000000000012', 'Hari sulit adalah latihan ketahanan; kamu sedang menguatkan mentalmu.', true, now()),
  ('30000000-0000-0000-0000-000000000013', 'Ingat alasanmu memulai: kesehatan, relasi, dan masa depan lebih baik.', true, now()),
  ('30000000-0000-0000-0000-000000000014', 'Recovery adalah bentuk kepedulian pada diri sendiri.', true, now()),
  ('30000000-0000-0000-0000-000000000015', 'Kemenangan kecil hari ini membentuk kemenangan besar nanti.', true, now()),
  ('30000000-0000-0000-0000-000000000016', 'Masa lalu tidak mendefinisikan masa depanmu.', true, now()),
  ('30000000-0000-0000-0000-000000000017', 'Urge akan berlalu; yang penting adalah responmu saat itu datang.', true, now()),
  ('30000000-0000-0000-0000-000000000018', 'Konsistensi membangun fondasi yang tidak mudah runtuh.', true, now()),
  ('30000000-0000-0000-0000-000000000019', 'Jangan bandingkan perjalananmu dengan orang lain.', true, now()),
  ('30000000-0000-0000-0000-000000000020', 'Saat ingin menyerah, ingat mengapa kamu mulai.', true, now()),
  ('30000000-0000-0000-0000-000000000021', 'Perjuanganmu hari ini sedang membentuk ketangguhanmu besok.', true, now()),
  ('30000000-0000-0000-0000-000000000022', 'Tetap hadir, tetap sadar, tetap memilih hal yang sehat.', true, now()),
  ('30000000-0000-0000-0000-000000000023', 'Pikiran negatif bukan perintah; kamu bisa memilih respon baru.', true, now()),
  ('30000000-0000-0000-0000-000000000024', 'Tidak perlu sempurna untuk terus bergerak maju.', true, now()),
  ('30000000-0000-0000-0000-000000000025', 'Investasi terbaik adalah menjaga kesehatan mentalmu setiap hari.', true, now()),
  ('30000000-0000-0000-0000-000000000026', 'Support systemmu melihat perjuanganmu dan itu berarti.', true, now()),
  ('30000000-0000-0000-0000-000000000027', 'Recovery adalah maraton, bukan sprint. Pelan tapi konsisten.', true, now()),
  ('30000000-0000-0000-0000-000000000028', 'Kamu memiliki kemampuan untuk melewati fase sulit ini.', true, now()),
  ('30000000-0000-0000-0000-000000000029', 'Setiap kali bertahan dari godaan, kamu membangun kendali diri.', true, now()),
  ('30000000-0000-0000-0000-000000000030', 'Hari ini mungkin berat, tetapi kamu tetap bisa memilih langkah sehat.', true, now()),
  ('30000000-0000-0000-0000-000000000031', 'Ritme kecil yang konsisten lebih kuat daripada motivasi sesaat.', true, now()),
  ('30000000-0000-0000-0000-000000000032', 'Satu keputusan benar pada momen sulit bisa mengubah arah harimu.', true, now()),
  ('30000000-0000-0000-0000-000000000033', 'Jika gagal hari ini, reset malam ini, mulai lagi besok pagi.', true, now()),
  ('30000000-0000-0000-0000-000000000034', 'Kamu bertumbuh setiap kali memilih sadar, bukan otomatis.', true, now()),
  ('30000000-0000-0000-0000-000000000035', 'Berani minta bantuan adalah tanda kekuatan, bukan kelemahan.', true, now())
ON CONFLICT (id) DO UPDATE
SET
  content = EXCLUDED.content,
  is_active = EXCLUDED.is_active;

-- 12) Daily challenges
INSERT INTO daily_challenges (id, title, description, content, is_active, created_at)
SELECT
  seed.id::uuid,
  'Tantangan Harian',
  seed.content,
  seed.content,
  seed.is_active,
  seed.created_at
FROM (
VALUES
  ('40000000-0000-0000-0000-000000000001', 'Bangun lebih awal dan mulai hari dengan rencana 3 prioritas sehat.', true, now()),
  ('40000000-0000-0000-0000-000000000002', 'Baca konten edukatif pemulihan minimal 15 menit.', true, now()),
  ('40000000-0000-0000-0000-000000000003', 'Olahraga 30 menit untuk menurunkan impuls dan stres.', true, now()),
  ('40000000-0000-0000-0000-000000000004', 'Lakukan cold shower atau teknik reset fisik cepat.', true, now()),
  ('40000000-0000-0000-0000-000000000005', 'Hubungi accountability partner dan kirim update singkat.', true, now()),
  ('40000000-0000-0000-0000-000000000006', 'Tonton atau dengar 1 materi pemulihan yang aplikatif.', true, now()),
  ('40000000-0000-0000-0000-000000000007', 'Tulis 5 hal yang disyukuri dan 3 trigger utama hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000008', 'Aktifkan pembatasan aplikasi pemicu untuk 24 jam.', true, now()),
  ('40000000-0000-0000-0000-000000000009', 'Tulis alasan personal mengapa kamu memilih hidup yang lebih sehat.', true, now()),
  ('40000000-0000-0000-0000-000000000010', 'Rapikan ruang kerja atau kamar selama 15 menit.', true, now()),
  ('40000000-0000-0000-0000-000000000011', 'Latihan mindful breathing 10 menit saat dorongan muncul.', true, now()),
  ('40000000-0000-0000-0000-000000000012', 'Lakukan satu tindakan kebaikan untuk orang lain hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000013', 'Identifikasi 3 trigger terbesar dan susun respon cadangannya.', true, now()),
  ('40000000-0000-0000-0000-000000000014', 'Luangkan waktu berkualitas tanpa gadget bersama keluarga atau teman.', true, now()),
  ('40000000-0000-0000-0000-000000000015', 'Jaga pola makan bersih dan hindari konsumsi pemicu berlebih.', true, now()),
  ('40000000-0000-0000-0000-000000000016', 'Journaling refleksi: progres, struggle, dan rencana besok.', true, now()),
  ('40000000-0000-0000-0000-000000000017', 'Tetapkan 3 target realistis untuk 7 hari ke depan.', true, now()),
  ('40000000-0000-0000-0000-000000000018', 'Dengarkan podcast self-improvement selama aktivitas rutin.', true, now()),
  ('40000000-0000-0000-0000-000000000019', 'Meditasi 10-15 menit untuk menurunkan reaktivitas.', true, now()),
  ('40000000-0000-0000-0000-000000000020', 'Buat vision board sederhana tentang tujuan hidup sehatmu.', true, now()),
  ('40000000-0000-0000-0000-000000000021', 'No social media selama 1 hari penuh.', true, now()),
  ('40000000-0000-0000-0000-000000000022', 'Ikut aktivitas komunitas positif offline atau online.', true, now()),
  ('40000000-0000-0000-0000-000000000023', 'Minum air cukup dan jaga energi tubuh sepanjang hari.', true, now()),
  ('40000000-0000-0000-0000-000000000024', 'Buat playlist fokus untuk dipakai saat urge meningkat.', true, now()),
  ('40000000-0000-0000-0000-000000000025', 'Saat urge muncul: napas 5 menit + 20 push-up + ganti lokasi.', true, now()),
  ('40000000-0000-0000-0000-000000000026', 'Hubungi kembali satu relasi positif yang lama tidak disapa.', true, now()),
  ('40000000-0000-0000-0000-000000000027', 'Catat semua kemenangan kecil sejak mulai recovery.', true, now()),
  ('40000000-0000-0000-0000-000000000028', 'Tidur lebih awal minimal 30 menit dari kebiasaan.', true, now()),
  ('40000000-0000-0000-0000-000000000029', 'Ubah satu self-talk negatif jadi afirmasi realistis.', true, now()),
  ('40000000-0000-0000-0000-000000000030', 'Luangkan 30 menit outdoor untuk reset pikiran.', true, now()),
  ('40000000-0000-0000-0000-000000000031', 'Audit histori browser dan bersihkan sumber pemicu.', true, now()),
  ('40000000-0000-0000-0000-000000000032', 'Gunakan mode fokus 2 sesi x 45 menit untuk tugas utama.', true, now()),
  ('40000000-0000-0000-0000-000000000033', 'Lakukan check-in emosi 3 kali: pagi, siang, malam.', true, now()),
  ('40000000-0000-0000-0000-000000000034', 'Siapkan emergency note di ponsel untuk dibaca saat krisis.', true, now()),
  ('40000000-0000-0000-0000-000000000035', 'Tutup hari dengan evaluasi: apa pemicu, apa respon, apa pelajaran.', true, now()),
  ('40000000-0000-0000-0000-000000000036', 'Lakukan digital sunset: hentikan layar 60 menit sebelum tidur.', true, now()),
  ('40000000-0000-0000-0000-000000000037', 'Buat daftar 5 aktivitas pengganti saat urge muncul.', true, now()),
  ('40000000-0000-0000-0000-000000000038', 'Praktikkan teknik grounding 5-4-3-2-1 minimal 2 kali hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000039', 'Lakukan jalan kaki 20 menit tanpa ponsel untuk reset pikiran.', true, now()),
  ('40000000-0000-0000-0000-000000000040', 'Bersihkan feed media sosial dari akun atau kata kunci pemicu.', true, now()),
  ('40000000-0000-0000-0000-000000000041', 'Tulis jurnal 10 menit tentang emosi yang paling dominan hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000042', 'Gunakan teknik pomodoro 3 sesi untuk tugas penting hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000043', 'Siapkan rencana besok malam ini: 3 prioritas dan 1 batasan digital.', true, now()),
  ('40000000-0000-0000-0000-000000000044', 'Hubungi satu orang support system dan kirim update progres singkat.', true, now()),
  ('40000000-0000-0000-0000-000000000045', 'Praktik self-talk sehat: ubah 3 pikiran negatif jadi kalimat realistis.', true, now()),
  ('40000000-0000-0000-0000-000000000046', 'Lakukan latihan napas 4-7-8 selama 5 menit saat stres meningkat.', true, now()),
  ('40000000-0000-0000-0000-000000000047', 'Atur ulang kamar/ruang kerja agar lebih minim distraksi visual.', true, now()),
  ('40000000-0000-0000-0000-000000000048', 'Buat checklist rutinitas pagi dan jalankan minimal 80%.', true, now()),
  ('40000000-0000-0000-0000-000000000049', 'Selesaikan 1 tugas yang paling sering ditunda (minimal 25 menit).', true, now()),
  ('40000000-0000-0000-0000-000000000050', 'Tutup hari dengan gratitude: tulis 3 hal baik yang terjadi hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000051', 'No gadget 30 menit setelah bangun agar pagi lebih terarah.', true, now()),
  ('40000000-0000-0000-0000-000000000052', 'Siapkan satu makanan sehat hari ini untuk menjaga energi stabil.', true, now()),
  ('40000000-0000-0000-0000-000000000053', 'Tuliskan 3 trigger utama minggu ini dan 1 respon pengganti untuk tiap trigger.', true, now()),
  ('40000000-0000-0000-0000-000000000054', 'Lakukan peregangan 10 menit setiap kali pikiran mulai buntu.', true, now()),
  ('40000000-0000-0000-0000-000000000055', 'Matikan notifikasi non-penting selama jam fokus utama.', true, now()),
  ('40000000-0000-0000-0000-000000000056', 'Baca ulang alasan recovery pribadimu sebelum tidur malam ini.', true, now()),
  ('40000000-0000-0000-0000-000000000057', 'Lakukan 1 aksi sosial positif: menyapa, membantu, atau memberi apresiasi.', true, now()),
  ('40000000-0000-0000-0000-000000000058', 'Gunakan rule 10 menit: mulai tugas sulit minimal 10 menit tanpa jeda.', true, now()),
  ('40000000-0000-0000-0000-000000000059', 'Selesaikan hari ini tanpa membuka konten random di jam rawan.', true, now()),
  ('40000000-0000-0000-0000-000000000060', 'Tulis rencana emergency 3 langkah untuk menghadapi urge mendadak.', true, now()),
  ('40000000-0000-0000-0000-000000000061', 'Lakukan check-in emosi tiap 4 jam dan catat perubahan mood.', true, now()),
  ('40000000-0000-0000-0000-000000000062', 'Luangkan 20 menit belajar skill baru sebagai pengganti scrolling.', true, now()),
  ('40000000-0000-0000-0000-000000000063', 'Rapikan file/folder digital untuk mengurangi distraksi visual.', true, now()),
  ('40000000-0000-0000-0000-000000000064', 'Jaga hidrasi: minum air teratur sepanjang hari minimal 8 gelas.', true, now()),
  ('40000000-0000-0000-0000-000000000065', 'Lakukan evaluasi sore: apa yang memicu, apa yang berhasil, apa yang diperbaiki.', true, now()),
  ('40000000-0000-0000-0000-000000000066', 'Pasang batas aplikasi hiburan maksimal 45 menit hari ini.', true, now()),
  ('40000000-0000-0000-0000-000000000067', 'Jadwalkan aktivitas malam yang menenangkan selain layar.', true, now()),
  ('40000000-0000-0000-0000-000000000068', 'Tuliskan 1 kemenangan kecil hari ini lalu bagikan ke accountability partner.', true, now()),
  ('40000000-0000-0000-0000-000000000069', 'Saat dorongan naik: berhenti, bernapas, pindah lokasi, lalu mulai aktivitas pengganti.', true, now()),
  ('40000000-0000-0000-0000-000000000070', 'Akhiri hari dengan review 5 menit dan niat spesifik untuk esok pagi.', true, now())
) AS seed (id, content, is_active, created_at)
ON CONFLICT (id) DO UPDATE
SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  content = EXCLUDED.content,
  is_active = EXCLUDED.is_active;

-- 13) Daily physical challenges
INSERT INTO daily_physical_challenges (id, title, description, is_active, created_at)
VALUES
  ('41000000-0000-0000-0000-000000000001', 'Aktivasi Pagi', 'Lakukan peregangan seluruh tubuh selama 10 menit setelah bangun tidur.', true, now()),
  ('41000000-0000-0000-0000-000000000002', 'Jalan Cepat', 'Jalan cepat 20 menit tanpa ponsel untuk reset energi dan fokus.', true, now()),
  ('41000000-0000-0000-0000-000000000003', 'Push-up Set', 'Selesaikan 3 set push-up (masing-masing 10 repetisi) dengan jeda terkontrol.', true, now()),
  ('41000000-0000-0000-0000-000000000004', 'Squat Set', 'Selesaikan 3 set squat (masing-masing 15 repetisi) untuk aktivasi kaki.', true, now()),
  ('41000000-0000-0000-0000-000000000005', 'Core Challenge', 'Tahan plank total 3 menit (boleh dibagi beberapa set).', true, now()),
  ('41000000-0000-0000-0000-000000000006', 'Mobility Break', 'Ambil 3 jeda mobilitas @5 menit sepanjang hari kerja.', true, now()),
  ('41000000-0000-0000-0000-000000000007', 'Cardio Ringan', 'Lakukan jogging atau sepeda statis minimal 15 menit dengan intensitas ringan.', true, now()),
  ('41000000-0000-0000-0000-000000000008', 'Tangga Aktif', 'Naik turun tangga total 10 menit sebagai pengganti duduk lama.', true, now()),
  ('41000000-0000-0000-0000-000000000009', 'Cold Finish', 'Akhiri mandi dengan 60 detik air dingin untuk melatih kontrol impuls.', true, now()),
  ('41000000-0000-0000-0000-000000000010', 'Breath + Body', 'Kombinasikan 5 menit napas terarah lalu 20 burpee ringan.', true, now()),
  ('41000000-0000-0000-0000-000000000011', 'Sunlight Walk', 'Kena paparan sinar matahari pagi sambil jalan santai 15 menit.', true, now()),
  ('41000000-0000-0000-0000-000000000012', 'Desk Reset', 'Setiap 60 menit duduk, lakukan 1 menit gerak aktif (stretch atau squat).', true, now()),
  ('41000000-0000-0000-0000-000000000013', 'Evening Stretch', 'Lakukan routine peregangan malam 12 menit sebelum tidur.', true, now()),
  ('41000000-0000-0000-0000-000000000014', 'Glute Bridge', 'Kerjakan 3 set glute bridge (masing-masing 15 repetisi).', true, now()),
  ('41000000-0000-0000-0000-000000000015', 'Wall Sit', 'Tahan wall sit total 2 menit, dibagi menjadi 2-4 set.', true, now()),
  ('41000000-0000-0000-0000-000000000016', 'Lunge Flow', 'Lakukan alternating lunge total 24 repetisi dengan gerakan terkontrol.', true, now()),
  ('41000000-0000-0000-0000-000000000017', 'Low Impact HIIT', 'Selesaikan 10 menit low-impact HIIT (work 30 detik, rest 30 detik).', true, now()),
  ('41000000-0000-0000-0000-000000000018', 'Hip Mobility', 'Lakukan latihan mobilitas pinggul 8-10 menit untuk kurangi ketegangan.', true, now()),
  ('41000000-0000-0000-0000-000000000019', 'Shoulder Release', 'Kerjakan drill bahu dan punggung atas 10 menit untuk postur lebih baik.', true, now()),
  ('41000000-0000-0000-0000-000000000020', 'Night Walk', 'Jalan santai 15 menit setelah makan malam untuk menurunkan stres.', true, now())
ON CONFLICT (id) DO UPDATE
SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  is_active = EXCLUDED.is_active;

-- 14) Achievements catalog
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
  ('77777777-7777-7777-7777-777777777777', 'streak_7_days', '7 Hari Konsisten', 'Pertahankan streak berhasil selama 7 hari berturut-turut.', 'streak_milestone', 7, true, now(), now()),
  ('88888888-8888-8888-8888-888888888888', 'checkin_20_of_30', 'Check-in Konsisten', 'Selesaikan minimal 20 check-in berhasil dalam 30 hari terakhir.', 'checkin_consistency', 20, true, now(), now()),
  ('99999999-9999-9999-9999-999999999999', 'journal_15_of_30', 'Jurnal Reflektif', 'Tulis minimal 15 jurnal dalam 30 hari terakhir.', 'journal_consistency', 15, true, now(), now()),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'relapse_recovery_3', 'Bangkit Setelah Relapse', 'Capai 3 check-in berhasil setelah relapse terakhir.', 'relapse_recovery', 3, true, now(), now()),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'community_participation_10', 'Aktif di Komunitas', 'Capai total 10 interaksi komunitas (post, komentar, atau like).', 'community_participation', 10, true, now(), now()),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'onboarding_complete', 'Onboarding Selesai', 'Selesaikan onboarding profil pemulihan.', 'onboarding_completion', 1, true, now(), now()),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'streak_30_days', '30 Hari Beruntun', 'Pertahankan streak berhasil selama 30 hari berturut-turut.', 'streak_milestone', 30, true, now(), now()),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'checkin_60_of_90', 'Ketahanan 90 Hari', 'Selesaikan minimal 60 check-in berhasil dalam 90 hari terakhir.', 'checkin_consistency', 60, true, now(), now()),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'journal_30_of_60', 'Ritme Jurnal Stabil', 'Tulis minimal 30 jurnal dalam 60 hari terakhir.', 'journal_consistency', 30, true, now(), now()),
  ('abababab-abab-abab-abab-abababababab', 'community_participation_25', 'Kolaborator Komunitas', 'Capai total 25 interaksi komunitas yang konstruktif.', 'community_participation', 25, true, now(), now()),
  ('acacacac-acac-acac-acac-acacacacacac', 'streak_90_days', '90 Hari Tangguh', 'Pertahankan streak berhasil selama 90 hari berturut-turut.', 'streak_milestone', 90, true, now(), now()),
  ('adadadad-adad-adad-adad-adadadadadad', 'checkin_120_of_180', 'Disiplin 180 Hari', 'Selesaikan minimal 120 check-in berhasil dalam 180 hari terakhir.', 'checkin_consistency', 120, true, now(), now()),
  ('aeaeaeae-aeae-aeae-aeae-aeaeaeaeaeae', 'journal_60_of_120', 'Penulis Konsisten', 'Tulis minimal 60 jurnal dalam 120 hari terakhir.', 'journal_consistency', 60, true, now(), now()),
  ('afafafaf-afaf-afaf-afaf-afafafafafaf', 'community_support_40', 'Pilar Komunitas', 'Capai 40 interaksi komunitas bermakna.', 'community_participation', 40, true, now(), now()),
  ('b0b0b0b0-b0b0-b0b0-b0b0-b0b0b0b0b0b0', 'ai_reflection_20', 'Refleksi AI Aktif', 'Selesaikan minimal 20 sesi refleksi dengan AI coach.', 'ai_engagement', 20, true, now(), now())
ON CONFLICT (code) DO UPDATE
SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  threshold = EXCLUDED.threshold,
  is_active = EXCLUDED.is_active,
  updated_at = now();

-- 15) Achievement progress
WITH progress_seed (user_id, achievement_code, progress_value, unlocked_at, evaluated_at) AS (
  VALUES
    ('10000000-0000-0000-0000-000000000001', 'streak_7_days', 14, now() - interval '20 days', now()),
    ('10000000-0000-0000-0000-000000000001', 'checkin_20_of_30', 24, now() - interval '5 days', now()),
    ('10000000-0000-0000-0000-000000000001', 'journal_15_of_30', 18, now() - interval '3 days', now()),
    ('10000000-0000-0000-0000-000000000001', 'community_participation_10', 11, now() - interval '2 days', now()),
    ('10000000-0000-0000-0000-000000000002', 'streak_7_days', 21, now() - interval '18 days', now()),
    ('10000000-0000-0000-0000-000000000002', 'checkin_20_of_30', 27, now() - interval '4 days', now()),
    ('10000000-0000-0000-0000-000000000002', 'community_participation_10', 13, now() - interval '4 days', now()),
    ('10000000-0000-0000-0000-000000000002', 'community_participation_25', 19, NULL, now()),
    ('10000000-0000-0000-0000-000000000003', 'streak_7_days', 9, now() - interval '14 days', now()),
    ('10000000-0000-0000-0000-000000000003', 'journal_15_of_30', 14, NULL, now()),
    ('10000000-0000-0000-0000-000000000003', 'relapse_recovery_3', 2, NULL, now()),
    ('10000000-0000-0000-0000-000000000003', 'ai_reflection_20', 8, NULL, now()),
    ('10000000-0000-0000-0000-000000000004', 'streak_7_days', 8, now() - interval '10 days', now()),
    ('10000000-0000-0000-0000-000000000004', 'relapse_recovery_3', 3, now() - interval '6 days', now()),
    ('10000000-0000-0000-0000-000000000004', 'community_participation_10', 7, NULL, now()),
    ('10000000-0000-0000-0000-000000000004', 'ai_reflection_20', 12, NULL, now()),
    ('10000000-0000-0000-0000-000000000005', 'streak_7_days', 28, now() - interval '22 days', now()),
    ('10000000-0000-0000-0000-000000000005', 'checkin_20_of_30', 26, now() - interval '7 days', now()),
    ('10000000-0000-0000-0000-000000000005', 'journal_15_of_30', 16, now() - interval '7 days', now()),
    ('10000000-0000-0000-0000-000000000005', 'onboarding_complete', 1, now() - interval '105 days', now()),
    ('10000000-0000-0000-0000-000000000006', 'streak_7_days', 7, now() - interval '3 days', now()),
    ('10000000-0000-0000-0000-000000000006', 'community_participation_10', 10, now() - interval '1 days', now()),
    ('10000000-0000-0000-0000-000000000006', 'ai_reflection_20', 15, NULL, now()),
    ('10000000-0000-0000-0000-000000000006', 'onboarding_complete', 1, now() - interval '102 days', now())
)
INSERT INTO user_achievement_progress (
  id,
  user_id,
  achievement_id,
  progress_value,
  unlocked_at,
  last_evaluated_at,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  ps.user_id::uuid,
  a.id,
  ps.progress_value,
  ps.unlocked_at,
  ps.evaluated_at,
  now() - interval '30 days',
  now()
FROM progress_seed ps
JOIN achievements a ON a.code = ps.achievement_code
ON CONFLICT (user_id, achievement_id) DO UPDATE
SET
  progress_value = EXCLUDED.progress_value,
  unlocked_at = EXCLUDED.unlocked_at,
  last_evaluated_at = EXCLUDED.last_evaluated_at,
  updated_at = now();

-- 16) AI persona preferences
INSERT INTO user_ai_persona_preferences (
  user_id,
  persona,
  updated_at
)
VALUES
  ('10000000-0000-0000-0000-000000000001', 'supportive', now()),
  ('10000000-0000-0000-0000-000000000002', 'friendly', now()),
  ('10000000-0000-0000-0000-000000000003', 'direct', now()),
  ('10000000-0000-0000-0000-000000000004', 'supportive', now()),
  ('10000000-0000-0000-0000-000000000005', 'concise', now()),
  ('10000000-0000-0000-0000-000000000006', 'friendly', now())
ON CONFLICT (user_id) DO UPDATE
SET
  persona = EXCLUDED.persona,
  updated_at = now();

-- 17) AI chats
INSERT INTO ai_chats (
  id,
  user_id,
  role,
  content,
  created_at
)
VALUES
  ('90000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'user', 'Hari ini dorongan muncul setelah lembur, saya butuh langkah cepat.', now() - interval '6 days'),
  ('90000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 'assistant', 'Ambil jeda 10 menit: pindah lokasi, minum air, lalu lakukan napas 4-7-8 tiga putaran.', now() - interval '6 days' + interval '1 minute'),
  ('90000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 'user', 'Baik, saya lanjut dengan jalan kaki singkat sekarang.', now() - interval '6 days' + interval '2 minutes'),
  ('90000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', 'user', 'Saya ingin evaluasi kebiasaan malam supaya tidak relapse.', now() - interval '5 days'),
  ('90000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000002', 'assistant', 'Mulai dari cutoff layar jam 21.00 dan journaling 5 menit sebelum tidur.', now() - interval '5 days' + interval '1 minute'),
  ('90000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000002', 'user', 'Saya set alarm pengingat malam ini.', now() - interval '5 days' + interval '2 minutes'),
  ('90000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000003', 'user', 'Brain fog saya masih tinggi, sulit fokus kerja.', now() - interval '4 days'),
  ('90000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000003', 'assistant', 'Coba blok kerja 25 menit + jeda 5 menit, dan prioritaskan satu tugas inti dahulu.', now() - interval '4 days' + interval '1 minute'),
  ('90000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000003', 'user', 'Saya mulai dari tugas paling penting dulu.', now() - interval '4 days' + interval '2 minutes'),
  ('90000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000004', 'user', 'Malam ini cemas dan takut mengulang pola lama.', now() - interval '3 days'),
  ('90000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000004', 'assistant', 'Pakai protokol darurat: napas, grounding, lalu hubungi partner dukungan.', now() - interval '3 days' + interval '1 minute'),
  ('90000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000004', 'user', 'Sudah kirim pesan ke support system, kondisi lebih tenang.', now() - interval '3 days' + interval '2 minutes'),
  ('90000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000005', 'user', 'Deadline kampus bikin saya ingin lari ke distraksi.', now() - interval '2 days'),
  ('90000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000005', 'assistant', 'Pecah tugas jadi 3 blok kecil dan beri jeda aktif tiap blok.', now() - interval '2 days' + interval '1 minute'),
  ('90000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000005', 'user', 'Oke, saya mulai blok pertama sekarang.', now() - interval '2 days' + interval '2 minutes'),
  ('90000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000006', 'user', 'Saya ingin menjaga konsistensi refleksi harian.', now() - interval '1 day'),
  ('90000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000006', 'assistant', 'Gunakan format 3 pertanyaan: trigger hari ini, respon, pelajaran besok.', now() - interval '1 day' + interval '1 minute'),
  ('90000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000006', 'user', 'Format ini sederhana dan mudah dijalankan.', now() - interval '1 day' + interval '2 minutes')
ON CONFLICT (id) DO UPDATE
SET
  role = EXCLUDED.role,
  content = EXCLUDED.content,
  created_at = EXCLUDED.created_at;

-- 18) Relapses (dari check-in gagal 14 hari terakhir)
INSERT INTO relapses (
  id,
  user_id,
  check_in_id,
  relapse_date,
  mood,
  commitment,
  relapse_trigger,
  created_at
)
SELECT
  gen_random_uuid(),
  ci.user_id,
  ci.id,
  ci.check_in_date,
  left(ci.mood, 50),
  format(
    'Komitmen reset %s: batasi akses pemicu, hubungi support system, dan jalankan emergency routine 15 menit.',
    to_char(ci.check_in_date, 'YYYY-MM-DD')
  ),
  COALESCE(ci.relapse_trigger, ARRAY['pemicu belum tercatat']::text[]),
  ci.check_in_date::timestamp + interval '22 hours'
FROM check_ins ci
WHERE ci.user_id IN (
  '10000000-0000-0000-0000-000000000001',
  '10000000-0000-0000-0000-000000000002',
  '10000000-0000-0000-0000-000000000003',
  '10000000-0000-0000-0000-000000000004',
  '10000000-0000-0000-0000-000000000005',
  '10000000-0000-0000-0000-000000000006'
)
  AND ci.check_in_date BETWEEN current_date - 13 AND current_date
  AND ci.is_successful = false
ON CONFLICT (user_id, relapse_date) DO UPDATE
SET
  check_in_id = EXCLUDED.check_in_id,
  mood = EXCLUDED.mood,
  commitment = EXCLUDED.commitment,
  relapse_trigger = EXCLUDED.relapse_trigger,
  created_at = EXCLUDED.created_at;

-- 19) Auth refresh tokens (hash-only)
INSERT INTO auth_refresh_tokens (
  id,
  user_id,
  token_hash,
  expires_at,
  revoked_at,
  rotated_from_id,
  created_at
)
VALUES
  ('a1000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'sha256:seed-user1-active-v1', now() + interval '20 days', NULL, NULL, now() - interval '2 days'),
  ('a1000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 'sha256:seed-user2-rotated-v1', now() + interval '15 days', now() - interval '5 days', NULL, now() - interval '20 days'),
  ('a1000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 'sha256:seed-user2-active-v2', now() + interval '25 days', NULL, 'a1000000-0000-0000-0000-000000000002', now() - interval '5 days'),
  ('a1000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000003', 'sha256:seed-user3-expired-v1', now() - interval '1 day', NULL, NULL, now() - interval '30 days'),
  ('a1000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000004', 'sha256:seed-user4-revoked-v1', now() + interval '10 days', now() - interval '1 day', NULL, now() - interval '12 days'),
  ('a1000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000005', 'sha256:seed-user5-active-v1', now() + interval '30 days', NULL, NULL, now() - interval '3 days'),
  ('a1000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000006', 'sha256:seed-user6-active-v1', now() + interval '18 days', NULL, NULL, now() - interval '4 days')
ON CONFLICT (id) DO UPDATE
SET
  token_hash = EXCLUDED.token_hash,
  expires_at = EXCLUDED.expires_at,
  revoked_at = EXCLUDED.revoked_at,
  rotated_from_id = EXCLUDED.rotated_from_id,
  created_at = EXCLUDED.created_at;

COMMIT;
