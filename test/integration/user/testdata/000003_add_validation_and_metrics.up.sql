-- เพิ่มตาราง migration_metrics
DROP TABLE IF EXISTS migration_metrics;

CREATE TABLE migration_metrics (
    id INT AUTO_INCREMENT PRIMARY KEY,
    migration_version VARCHAR(50),
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    records_processed INT,
    errors_count INT,
    status ENUM('running', 'completed', 'failed'),
    error_message TEXT
);

-- Procedure สำหรับ validate user data
DROP PROCEDURE IF EXISTS validate_user_data;

CREATE PROCEDURE validate_user_data()
BEGIN
    DECLARE invalid_emails INT;
    DECLARE invalid_usernames INT;
    DECLARE error_message TEXT;
    
    -- ตรวจสอบ email format
    SELECT COUNT(*) INTO invalid_emails
    FROM users 
    WHERE email NOT REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$';
    
    -- ตรวจสอบ username format
    SELECT COUNT(*) INTO invalid_usernames
    FROM users
    WHERE username NOT REGEXP '^[a-zA-Z0-9_-]{3,50}$';
    
    -- สร้าง error message ถ้าพบข้อมูลไม่ถูกต้อง
    IF invalid_emails > 0 OR invalid_usernames > 0 THEN
        SET error_message = CONCAT(
            'Data validation failed: ',
            IF(invalid_emails > 0, CONCAT(invalid_emails, ' invalid emails. '), ''),
            IF(invalid_usernames > 0, CONCAT(invalid_usernames, ' invalid usernames.'), '')
        );
        
        SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = error_message;
    END IF;
END;