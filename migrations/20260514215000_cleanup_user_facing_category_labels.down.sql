UPDATE education_contents
SET category = replace(category, ' ', '_')
WHERE category IN (
  'pola pikir',
  'regulasi emosi',
  'strategi harian',
  'kesehatan mental',
  'kesehatan holistik',
  'pengembangan diri',
  'kesadaran diri'
);
