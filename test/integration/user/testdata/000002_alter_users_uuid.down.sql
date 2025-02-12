-- สำรองข้อมูลและเปลี่ยนกลับเป็น BIGINT
CREATE TABLE users_temp (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    status ENUM('active', 'inactive', 'suspended') NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1
);

-- Copy ข้อมูลจากตารางเดิม โดยแปลง UUID เป็น BIGINT
INSERT INTO users_temp (
    username, 
    email,
    full_name,
    password,
    status,
    created_at,
    updated_at,
    version
)
SELECT 
    username,
    email,
    full_name,
    password,
    status,
    created_at,
    updated_at,
    version
FROM users;

-- ลบตารางเดิมและเปลี่ยนชื่อตารางใหม่
DROP TABLE users;
RENAME TABLE users_temp TO users;