SELECT 'CREATE DATABASE question_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'question_db')\gexec
