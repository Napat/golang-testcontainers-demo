# คู่มือการใช้งาน Database Migrations

## แนวคิดพื้นฐาน

Database migrations คือการจัดการการเปลี่ยนแปลงโครงสร้างฐานข้อมูลอย่างเป็นระบบ โดยมีหลักการสำคัญดังนี้:

1. การควบคุมเวอร์ชัน (Version Control)
   - แต่ละการเปลี่ยนแปลงจะมีเลขเวอร์ชันกำกับ
   - เก็บประวัติการเปลี่ยนแปลงทั้งหมด
   - สามารถย้อนกลับไปเวอร์ชันก่อนหน้าได้

2. การทำงานแบบสองทิศทาง (Bi-directional)
   - Up Migration: สำหรับการอัพเกรด
   - Down Migration: สำหรับการดาวน์เกรด
   - ต้องเขียนทั้งสองส่วนเสมอเพื่อให้ย้อนกลับได้

3. การรักษาข้อมูล (Data Preservation)
   - ต้องคำนึงถึงข้อมูลที่มีอยู่เดิม
   - มีขั้นตอนสำรองข้อมูลก่อนเปลี่ยนแปลง
   - มีการตรวจสอบความถูกต้องของข้อมูล

## โครงสร้างไฟล์ Migration

### รูปแบบการตั้งชื่อไฟล์

```sh
{version}_{description}.{direction}.sql
```

ตัวอย่าง:

- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`

### หลักการตั้งชื่อ

1. version: ตัวเลขเรียงลำดับ เติม 0 ข้างหน้า (เช่น 000001)
2. description: คำอธิบายสั้นๆ ว่าทำอะไร
3. direction: up หรือ down
4. นามสกุล: .sql

## เครื่องมือที่ใช้

### golang-migrate

เราใช้ `golang-migrate` เป็นเครื่องมือหลักในการจัดการ migrations เนื่องจาก:

- รองรับหลายฐานข้อมูล (MySQL, PostgreSQL)
- มี CLI ใช้งานง่าย
- ทำงานร่วมกับ Go ได้ดี
- มีระบบจัดการ versions

วิธีติดตั้ง:

```bash
go install -tags 'mysql postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## วิธีใช้งาน Scripts

### migrate.sh

คำสั่งพื้นฐาน:

```bash
# MySQL
./scripts/migrate.sh --type mysql --command up
./scripts/migrate.sh --type mysql --command down
./scripts/migrate.sh --type mysql --command version

# PostgreSQL
./scripts/migrate.sh --type postgres --command up
./scripts/migrate.sh --type postgres --command down
./scripts/migrate.sh --type postgres --command version
```

### migrate-to.sh

ใช้เมื่อต้องการย้ายไปยังเวอร์ชันที่ต้องการโดยตรง:

```bash
# MySQL
./scripts/migrate-to.sh --type mysql --version <target_version>

# PostgreSQL
./scripts/migrate-to.sh --type postgres --version <target_version>
```

พารามิเตอร์ที่รองรับ:

- `--type`: ระบุประเภทฐานข้อมูล (mysql หรือ postgres)
- `--user`: ชื่อผู้ใช้ (ค่าเริ่มต้น: root สำหรับ MySQL, postgres สำหรับ PostgreSQL)
- `--password`: รหัสผ่าน
- `--database`: ชื่อฐานข้อมูล
- `--path`: path ของไฟล์ migrations

## แนวทางการเขียน Migration ที่ดี

### 1. การเขียน Up Migration

- เขียนคำสั่งสร้างหรือแก้ไขโครงสร้าง
- ระวังเรื่องลำดับการทำงาน
- ใช้ IF NOT EXISTS เพื่อป้องกันความผิดพลาด
- เพิ่ม indexes ที่จำเป็น

ตัวอย่าง:

```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE
);

CREATE INDEX idx_users_username ON users(username);
```

### 2. การเขียน Down Migration

- ต้องสามารถย้อนกลับได้สมบูรณ์
- รักษาข้อมูลสำคัญไว้
- ทำ cleanup ให้เรียบร้อย

ตัวอย่าง:

```sql
-- สำรองข้อมูลก่อนลบ
CREATE TABLE users_backup AS SELECT * FROM users;
DROP TABLE users;
```

### 3. การตรวจสอบความถูกต้อง

- ใช้ stored procedures สำหรับตรวจสอบ
- เพิ่ม constraints ที่เหมาะสม
- ทดสอบทั้ง up และ down

## การแก้ไขปัญหาที่พบบ่อย

### 1. Version Conflicts

ตรวจสอบสถานะปัจจุบัน:

```sql
SELECT * FROM schema_migrations;
```

แก้ไขสถานะ dirty:

```sql
UPDATE schema_migrations 
SET dirty = false 
WHERE version = <current_version>;
```

### 2. ข้อมูลไม่สมบูรณ์

- ตรวจสอบ constraints
- ใช้ validation procedures
- แก้ไขข้อมูลให้ถูกต้อง

### 3. การเชื่อมต่อฐานข้อมูล

MySQL:

```sh
Host: localhost
Port: 3306
User: root
Password: password
Database: testdb
```

PostgreSQL:

```sh
Host: localhost
Port: 5432
User: postgres
Password: password
Database: testdb
```

## ระบบติดตามการทำงาน

ตาราง migration_metrics ใช้เก็บข้อมูลการทำงาน:

- เวอร์ชันที่ทำ migration
- เวลาเริ่มต้นและสิ้นสุด
- จำนวนข้อมูลที่ประมวลผล
- จำนวนข้อผิดพลาด
- สถานะการทำงาน

## คำแนะนำเพิ่มเติม

1. ทำการ backup ข้อมูลก่อนรัน migration เสมอ
2. ทดสอบใน development ก่อนรันจริง
3. เขียน migration ให้เป็น atomic
4. ตรวจสอบผลลัพธ์ทุกครั้ง
5. เก็บ log การทำงานไว้ตรวจสอบ

## ความแตกต่างระหว่าง MySQL และ PostgreSQL

### MySQL

- ใช้ ENUM type แบบ native
- ใช้ AUTO_INCREMENT
- มี ON UPDATE CURRENT_TIMESTAMP
- Container: testcontainers-demo-mysql-1

### PostgreSQL

- ต้องสร้าง ENUM ด้วย CREATE TYPE
- ใช้ SERIAL หรือ IDENTITY
- รองรับ JSONB
- Container: testcontainers-demo-postgres-1

## การเก็บ Metrics และการตรวจสอบ

### Migration Metrics

ระบบมีการเก็บ metrics ผ่านตาราง migration_metrics:

```sql
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
```

### การวิเคราะห์ Performance

1. ตรวจสอบระยะเวลาการทำงาน:

```sql
SELECT 
    migration_version,
    TIMESTAMPDIFF(SECOND, start_time, end_time) as duration_seconds,
    records_processed,
    status
FROM migration_metrics
ORDER BY start_time DESC;
```

2. ค้นหา migrations ที่มีปัญหา:

```sql
SELECT * FROM migration_metrics 
WHERE status = 'failed' 
OR errors_count > 0
ORDER BY start_time DESC;
```

## Zero Downtime Migration

### แนวทางการทำ Zero Downtime

1. **Backward Compatible Changes**
   - เพิ่มคอลัมน์ใหม่แบบ nullable ก่อน
   - ค่อยๆ update application code
   - ลบคอลัมน์เก่าในภายหลัง

2. **การแยก Read/Write**
   - ใช้ read replica สำหรับ queries
   - ทำ migration บน master
   - ค่อยๆ sync replicas

3. **Shadow Tables**

   ```sql
   -- 1. สร้างตารางใหม่
   CREATE TABLE users_new LIKE users;
   ALTER TABLE users_new ADD COLUMN new_field VARCHAR(255);
   
   -- 2. Copy และ sync ข้อมูล
   INSERT INTO users_new
   SELECT *, NULL as new_field
   FROM users;
   
   -- 3. เปลี่ยนชื่อตารางแบบ atomic
   RENAME TABLE users TO users_old,
                users_new TO users;
   ```

4. **การทำ Rollback**
   - เตรียม rollback plan
   - ทดสอบ rollback procedure
   - เก็บข้อมูลสำรอง

### ขั้นตอนการทำ Zero Downtime

1. **เตรียมการ**

   - ตรวจสอบ disk space
   - วางแผนช่วงเวลา
   - เตรียม monitoring

2. **การ Deploy**
   - Deploy แบบ rolling update
   - ตรวจสอบ health check
   - มี circuit breaker

3. **การ Rollback**
   - กำหนด rollback triggers
   - automated rollback
   - การกู้คืนข้อมูล

## การจัดการ Long-running Migrations

### การแบ่ง Batch

1. **Chunking Data**

   ```sql
   -- ตัวอย่างการ update ทีละ chunk
   DELIMITER //
   CREATE PROCEDURE update_in_batches()
   BEGIN
     DECLARE done INT DEFAULT FALSE;
     DECLARE batch_size INT DEFAULT 1000;
     DECLARE offset_count INT DEFAULT 0;
     
     batch_loop: LOOP
       UPDATE users
       SET status = 'active'
       WHERE id IN (
         SELECT id FROM users
         WHERE status IS NULL
         LIMIT batch_size OFFSET offset_count
       );
       
       IF ROW_COUNT() = 0 THEN
         LEAVE batch_loop;
       END IF;
       
       SET offset_count = offset_count + batch_size;
       
       -- บันทึก progress
       INSERT INTO migration_metrics (migration_version, records_processed)
       VALUES ('current_version', ROW_COUNT());
       
       -- หน่วงเวลาเพื่อลด load
       DO SLEEP(0.1);
     END LOOP;
   END //
   DELIMITER ;
   ```

2. **Progress Tracking**
   - บันทึก state ของแต่ละ batch
   - สามารถ resume ได้
   - แสดง progress realtime

3. **Resource Management**
   - จำกัด batch size
   - ควบคุม connection pool
   - มี timeout และ retry

### การจัดการ Background Jobs

1. **Queue System**

   ```sql
   CREATE TABLE migration_jobs (
       id INT AUTO_INCREMENT PRIMARY KEY,
       job_type VARCHAR(50),
       status ENUM('pending', 'running', 'completed', 'failed'),
       started_at TIMESTAMP,
       completed_at TIMESTAMP,
       error_message TEXT,
       retry_count INT DEFAULT 0,
       max_retries INT DEFAULT 3
   );
   ```

2. **Job Monitoring**
   - ตรวจสอบ stuck jobs
   - แจ้งเตือนเมื่อ error
   - cleanup completed jobs

## การตรวจสอบ Dependencies

### Schema Dependencies

1. **Foreign Key Graph**

   ```sql
   SELECT
       TABLE_NAME,
       COLUMN_NAME,
       REFERENCED_TABLE_NAME,
       REFERENCED_COLUMN_NAME
   FROM
       INFORMATION_SCHEMA.KEY_COLUMN_USAGE
   WHERE
       REFERENCED_TABLE_SCHEMA = DATABASE()
   ORDER BY
       TABLE_NAME,
       REFERENCED_TABLE_NAME;
   ```

2. **Index Dependencies**
   - ตรวจสอบ index ที่ใช้งาน
   - วิเคราะห์ query patterns
   - optimize index usage

### Data Dependencies

1. **Circular References**
   - ตรวจหา circular FK
   - แก้ไขด้วย intermediate tables
   - ใช้ deferrable constraints

2. **Soft Dependencies**
   - Business logic constraints
   - Application-level validations
   - Cache dependencies

### การจัดการ Dependencies

1. **Migration Order**
   - สร้าง dependency graph
   - จัดลำดับการ migrate
   - ทดสอบ dependencies

2. **Breaking Changes**
   - ระบุ breaking changes
   - วางแผนการ migrate
   - แจ้งทีมที่เกี่ยวข้อง

## แนวทางการทดสอบ Migrations

### Unit Tests

```go
func TestMigration_001(t *testing.T) {
    // เตรียม test database
    db := setupTestDB()
    
    // รัน migration
    err := runMigration("001_create_users")
    assert.NoError(t, err)
    
    // ตรวจสอบ schema
    assertTableExists(t, db, "users")
    assertColumnExists(t, db, "users", "email")
    
    // ทดสอบ data integrity
    insertTestData(t, db)
    validateConstraints(t, db)
    
    // ทดสอบ rollback
    err = rollbackMigration("001_create_users")
    assert.NoError(t, err)
}
```

### Integration Tests

1. **End-to-End Testing**
   - ทดสอบกับ real database
   - ตรวจสอบ application
   - วัด performance

2. **Load Testing**
   - ทดสอบ concurrent users
   - วัด response time
   - ตรวจสอบ resource usage

## Monitoring และ Alerting

### Key Metrics

1. **Performance Metrics**
   - Migration duration
   - Records processed/second
   - Error rate

2. **Resource Usage**
   - CPU/Memory usage
   - Disk I/O
   - Connection count

3. **Application Impact**
   - Response time
   - Error rate
   - Query performance

### Alert Rules

1. **Critical Alerts**
   - Migration failure
   - Excessive duration
   - Data corruption

2. **Warning Alerts**
   - High resource usage
   - Slow progress
   - Retry attempts

## Recovery Procedures

### Immediate Actions

1. Stop affected migrations
2. Assess data integrity
3. Switch to fallback system

### Recovery Steps

1. **Data Recovery**
   - Restore from backup
   - Fix corrupted data
   - Validate integrity

2. **System Recovery**
   - Restart services
   - Clear caches
   - Update configurations

3. **Verification**
   - Run health checks
   - Validate business logic
   - Monitor performance

## ตัวอย่างการใช้งาน Procedures

### การทำ Batch Update

```sql
-- อัพเดทข้อมูลครั้งละ 1000 records
-- ลองใหม่สูงสุด 3 ครั้งถ้าเกิด error
-- หน่วงเวลา 0.1 วินาทีระหว่าง batch
CALL batch_update_data(1000, 3, 0.1);

-- ตรวจสอบสถานะการทำงาน
SELECT * FROM migration_metrics 
WHERE migration_version = 'batch_update'
ORDER BY start_time DESC LIMIT 1;
```

### การตรวจสอบสถานะ Migrations

```sql
-- ดูภาพรวมของ migrations ทั้งหมด
CALL check_migration_status();

-- ผลลัพธ์จะแสดง:
-- 1. สรุปจำนวน migrations แต่ละสถานะ
-- 2. migrations ที่ทำงานนานเกิน 1 ชั่วโมง
-- 3. migrations ที่มีปัญหา 10 รายการล่าสุด
```

### การแก้ไขปัญหา Long-running Migrations

1. **ตรวจหา migrations ที่ทำงานค้าง**

   ```sql
   SELECT *
   FROM migration_metrics
   WHERE status = 'running'
   AND TIMESTAMPDIFF(HOUR, start_time, CURRENT_TIMESTAMP) > 1;
   ```

2. **ยกเลิก migration ที่ค้าง**

   ```sql
   UPDATE migration_metrics
   SET status = 'failed',
       end_time = CURRENT_TIMESTAMP,
       error_message = 'Cancelled due to timeout'
   WHERE id = <migration_id>;
   ```

3. **รันใหม่ด้วย batch size ที่เล็กลง**

   ```sql
   CALL batch_update_data(500, 3, 0.2);  -- ลดขนาด batch และเพิ่มเวลาพัก
   ```
