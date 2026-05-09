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
    'Understanding Triggers and Routines',
    'Core basics to identify daily triggers and build healthier responses.',
    'https://recova.app/education/memahami-trigger-dan-rutinitas',
    NULL,
    'mindset',
    true,
    now()
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'Grounding Technique 5-4-3-2-1',
    'A simple exercise to regain focus when urges appear.',
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
    'One healthy choice today still matters.',
    true,
    now()
  ),
  (
    '44444444-4444-4444-4444-444444444444',
    'Small consistent progress is stronger than short-lived intention.',
    true,
    now()
  )
ON CONFLICT (content) DO NOTHING;

INSERT INTO daily_challenges (id, content, is_active, created_at)
VALUES
  (
    '55555555-5555-5555-5555-555555555555',
    'Write down one main trigger today and your response plan.',
    true,
    now()
  ),
  (
    '66666666-6666-6666-6666-666666666666',
    'Take a 60-second pause before reacting when urges appear.',
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
    '7 Consistent Days',
    'Maintain a successful streak for 7 consecutive days.',
    'streak_milestone',
    7,
    true,
    now(),
    now()
  ),
  (
    '88888888-8888-8888-8888-888888888888',
    'checkin_20_of_30',
    'Consistent Check-in',
    'Complete at least 20 successful check-ins in the last 30 days.',
    'checkin_consistency',
    20,
    true,
    now(),
    now()
  ),
  (
    '99999999-9999-9999-9999-999999999999',
    'journal_15_of_30',
    'Reflective Journal',
    'Write at least 15 journal entries in the last 30 days.',
    'journal_consistency',
    15,
    true,
    now(),
    now()
  ),
  (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'relapse_recovery_3',
    'Recover After Relapse',
    'Achieve 3 successful check-ins after the latest relapse.',
    'relapse_recovery',
    3,
    true,
    now(),
    now()
  ),
  (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'community_participation_10',
    'Active in Community',
    'Reach a total of 10 community interactions (post, comment, or like).',
    'community_participation',
    10,
    true,
    now(),
    now()
  ),
  (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'onboarding_complete',
    'Onboarding Complete',
    'Complete recovery profile onboarding.',
    'onboarding_completion',
    1,
    true,
    now(),
    now()
  )
ON CONFLICT (code) DO NOTHING;
