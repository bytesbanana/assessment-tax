-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS config_id_seq;
-- Table Definition
DELETE TABLE if EXISTS "tax_configs";
CREATE TABLE "tax_configs" (
    "id" int4 NOT NULL DEFAULT nextval('config_id_seq'::regclass),
    "name" varchar(255) NOT NULL,
    "key" varchar(255) NOT NULL UNIQUE,
    "value" decimal(10, 2),
    "created_by" varchar,
    "created_at" timestamp DEFAULT now(),
    "updated_by" varchar,
    "updated_at" timestamp,
    PRIMARY KEY ("id")
);
INSERT INTO "tax_configs" (
        "name",
        "key",
        "value",
        "created_by"
    )
VALUES (
        'Personal tax deduction',
        'PERSONAL_DEDUCTION',
        60000,
        'system'
    ),
    (
        'Maximum K Receipt deduction',
        'MAX_K_RECEIPT_DEDUCTION',
        50000,
        'system'
    ),
    ;