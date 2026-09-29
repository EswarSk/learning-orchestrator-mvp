CREATE TABLE IF NOT EXISTS learning.course_generation_quota (
 user_id text NOT NULL,
 day date NOT NULL,
 attempts integer NOT NULL CHECK (attempts BETWEEN 1 AND 5),
 PRIMARY KEY(user_id,day)
);
