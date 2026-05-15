UPDATE education_contents
SET category = replace(category, '_', ' ')
WHERE category LIKE '%\_%' ESCAPE '\';
