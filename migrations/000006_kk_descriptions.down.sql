UPDATE books SET description_kk = '' WHERE description_kk != '';
UPDATE events SET title_kk = NULL, description_kk = NULL WHERE title_kk IS NOT NULL OR description_kk IS NOT NULL;
