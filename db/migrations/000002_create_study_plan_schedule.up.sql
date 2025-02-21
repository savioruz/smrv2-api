BEGIN;

-- Study Plans table
CREATE TABLE study_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    course_code VARCHAR(20) NOT NULL,
    class_code VARCHAR(10) NOT NULL,
    course_name VARCHAR(255) NOT NULL,
    credits INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    CONSTRAINT fk_study_plan_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- User Schedules table
CREATE TABLE user_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    study_plan_id UUID NOT NULL REFERENCES study_plans(id),
    course_code VARCHAR(20) NOT NULL,
    class_code VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    CONSTRAINT fk_schedule_study_plan
        FOREIGN KEY(study_plan_id) 
        REFERENCES study_plans(id)
        ON DELETE CASCADE
);

-- Scraped Schedules table
CREATE TABLE scraped_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_code VARCHAR(20) NOT NULL,
    class_code VARCHAR(20) NOT NULL,
    course_name VARCHAR(255) NOT NULL,
    credits INTEGER NOT NULL,
    day_of_week VARCHAR(20) NOT NULL,
    room_number VARCHAR(50) NOT NULL,
    semester VARCHAR(20) NOT NULL,
    start_time VARCHAR(5) NOT NULL,
    end_time VARCHAR(5) NOT NULL,
    lecturer_name TEXT DEFAULT NULL,
    study_program VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX idx_study_plans_user_id ON study_plans(user_id);
CREATE INDEX idx_study_plans_course_code ON study_plans(course_code);
CREATE INDEX idx_user_schedules_study_plan_id ON user_schedules(study_plan_id);
CREATE INDEX idx_scraped_schedules_course_code ON scraped_schedules(course_code);

CREATE INDEX idx_study_plans_deleted ON study_plans(deleted_at, user_id);
CREATE INDEX idx_user_schedules_deleted ON user_schedules(deleted_at, study_plan_id);
CREATE INDEX idx_scraped_schedules_deleted ON scraped_schedules(deleted_at, course_code);

ALTER TABLE scraped_schedules 
ADD CONSTRAINT unique_scraped_schedule 
UNIQUE (course_code, class_code, day_of_week, room_number, start_time);

COMMIT;
