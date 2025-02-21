BEGIN;

DROP TABLE study_plans;
DROP TABLE schedules;

DROP INDEX idx_study_plans_user_id;
DROP INDEX idx_study_plans_course_code;
DROP INDEX idx_user_schedules_study_plan_id;
DROP INDEX idx_scraped_schedules_course_code;

DROP INDEX idx_study_plans_deleted;
DROP INDEX idx_user_schedules_deleted;
DROP INDEX idx_scraped_schedules_deleted;

DROP CONSTRAINT unique_scraped_schedule;

COMMIT;