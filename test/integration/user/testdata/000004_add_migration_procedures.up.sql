-- Procedure สำหรับทำ batch update ข้อมูลขนาดใหญ่
DROP PROCEDURE IF EXISTS batch_update_data;

CREATE PROCEDURE batch_update_data(
    IN batch_size INT,
    IN max_retries INT,
    IN sleep_interval DECIMAL(5,2)
)
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE offset_count INT DEFAULT 0;
    DECLARE affected_rows INT;
    DECLARE retry_count INT DEFAULT 0;
    DECLARE migration_job_id INT;
    
    -- สร้าง migration job
    INSERT INTO migration_metrics (
        migration_version,
        start_time,
        status
    ) VALUES (
        'batch_update',
        CURRENT_TIMESTAMP,
        'running'
    );
    SET migration_job_id = LAST_INSERT_ID();
    
    batch_loop: LOOP
        -- ทำ batch update
        UPDATE users
        SET status = 'active',
            updated_at = CURRENT_TIMESTAMP
        WHERE id IN (
            SELECT id 
            FROM (
                SELECT id 
                FROM users 
                WHERE status IS NULL
                LIMIT batch_size OFFSET offset_count
            ) AS tmp
        );
        
        SET affected_rows = ROW_COUNT();
        
        -- อัพเดท metrics
        UPDATE migration_metrics
        SET records_processed = records_processed + affected_rows
        WHERE id = migration_job_id;
        
        -- ถ้าไม่มีข้อมูลให้อัพเดทแล้ว หรือ retry เกินกำหนด ให้จบการทำงาน
        IF affected_rows = 0 OR retry_count >= max_retries THEN
            UPDATE migration_metrics
            SET end_time = CURRENT_TIMESTAMP,
                status = IF(retry_count >= max_retries AND affected_rows > 0, 'failed', 'completed')
            WHERE id = migration_job_id;
            LEAVE batch_loop;
        END IF;
        
        -- เลื่อน offset ไปตาม batch size
        SET offset_count = offset_count + batch_size;
        
        -- หน่วงเวลาเพื่อลด server load
        DO SLEEP(sleep_interval);
    END LOOP;
    
    -- ถ้ามี error ให้บันทึก
    IF retry_count >= max_retries THEN
        UPDATE migration_metrics
        SET error_message = CONCAT('Failed after ', retry_count, ' retries'),
            errors_count = retry_count
        WHERE id = migration_job_id;
    END IF;
END;

-- Procedure สำหรับตรวจสอบ foreign key dependencies
DROP PROCEDURE IF EXISTS check_table_dependencies;

CREATE PROCEDURE check_table_dependencies(
    IN table_name VARCHAR(64)
)
BEGIN
    -- หา foreign keys ที่ reference ไปยังตารางอื่น
    SELECT 
        CONSTRAINT_NAME,
        TABLE_NAME,
        COLUMN_NAME,
        REFERENCED_TABLE_NAME,
        REFERENCED_COLUMN_NAME
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
    WHERE 
        TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = table_name
        AND REFERENCED_TABLE_NAME IS NOT NULL;
    
    -- หา foreign keys จากตารางอื่นที่ reference มาที่ตารางนี้
    SELECT 
        CONSTRAINT_NAME,
        TABLE_NAME,
        COLUMN_NAME,
        REFERENCED_TABLE_NAME,
        REFERENCED_COLUMN_NAME
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
    WHERE 
        TABLE_SCHEMA = DATABASE()
        AND REFERENCED_TABLE_NAME = table_name;
END;

-- Procedure สำหรับตรวจสอบ migration status
DROP PROCEDURE IF EXISTS check_migration_status;

CREATE PROCEDURE check_migration_status()
BEGIN
    -- สรุปสถานะ migrations ทั้งหมด
    SELECT 
        status,
        COUNT(*) as count,
        AVG(TIMESTAMPDIFF(SECOND, start_time, COALESCE(end_time, CURRENT_TIMESTAMP))) as avg_duration_seconds,
        SUM(records_processed) as total_records,
        SUM(errors_count) as total_errors
    FROM migration_metrics
    GROUP BY status;
    
    -- แสดง migrations ที่กำลังทำงานอยู่
    SELECT *
    FROM migration_metrics
    WHERE status = 'running'
    AND start_time < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 1 HOUR);
    
    -- แสดง migrations ที่มีปัญหา
    SELECT *
    FROM migration_metrics
    WHERE status = 'failed'
    OR errors_count > 0
    ORDER BY start_time DESC
    LIMIT 10;
END;