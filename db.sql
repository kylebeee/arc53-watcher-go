CREATE TABLE "provider" (
  "id" bigint unsigned NOT NULL,
  "type" enum('nfd') NOT NULL,
  "round" bigint unsigned NOT NULL,
  PRIMARY KEY ("id"),
  INDEX "type" ("type"),
  INDEX "round" ("round")
);

CREATE TABLE "provider_address" (
  "id" bigint unsigned NOT NULL,
  "address" varchar(58) NOT NULL,
  PRIMARY KEY ("id", "address")
);

CREATE TABLE "community" (
  "id" bigint unsigned NOT NULL,
  "version" varchar(6) NOT NULL,
  PRIMARY KEY ("id")
);

CREATE TABLE "community_json" (
  "id" bigint unsigned NOT NULL,
  "data" json NOT NULL,
  "malformed" tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY ("id")
);

CREATE TABLE "community_token" (
  "id" bigint unsigned NOT NULL,
  "asset_id" bigint unsigned NOT NULL,
  "image" varchar(256) DEFAULT NULL,
  "image_integrity" varchar(256) DEFAULT NULL,
  "image_mimetype" varchar(32) DEFAULT NULL,
  PRIMARY KEY ("id","asset_id"),
  UNIQUE KEY "indexed_asset" ("asset_id"),
  KEY "indexed_image" ("image")
);

CREATE TABLE "community_associate" (
  "id" bigint unsigned NOT NULL,
  "address" varchar(58) NOT NULL,
  "role" varchar(64) NOT NULL,
  "confirmed" tinyint(1) NOT NULL DEFAULT '0',
  "txn" varchar(64) DEFAULT NULL,
  PRIMARY KEY ("id","address"),
  KEY "address" ("address"),
  KEY "confirmed" ("confirmed")
);

CREATE TABLE "collection" (
  "id" varchar(24) NOT NULL,
  "provider_id" bigint unsigned NOT NULL,
  "name" varchar(128) NOT NULL,
  "description" text,
  "banner" bigint unsigned DEFAULT NULL,
  "avatar" bigint unsigned DEFAULT NULL,
  "network" varchar(128) DEFAULT NULL,
  "explicit" tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY ("id"),
  UNIQUE KEY "provider_id_name" ("provider_id","name"),
  KEY "provider_id" ("provider_id")
);

CREATE TABLE "collection_prefix" (
  "id" varchar(24) NOT NULL,
  "prefix" varchar(256) NOT NULL,
  PRIMARY KEY ("id","prefix"),
  KEY "id" ("id"),
  KEY "prefix" ("prefix")
);

CREATE TABLE "collection_address" (
  "id" varchar(24) NOT NULL,
  "address" varchar(58) NOT NULL,
  PRIMARY KEY ("id","address"),
  KEY "id" ("id"),
  KEY "address" ("address")
);

CREATE TABLE "collection_asset" (
  "id" varchar(24) NOT NULL,
  "asa_id" bigint unsigned NOT NULL,
  PRIMARY KEY ("id","asa_id")
);

CREATE TABLE "collection_excluded_asset" (
  "id" varchar(24) NOT NULL,
  "asa_id" bigint unsigned NOT NULL,
  PRIMARY KEY ("id","asa_id")
);

CREATE TABLE "collection_artist" (
  "id" varchar(24) NOT NULL,
  "address" varchar(58) NOT NULL,
  PRIMARY KEY ("id","address"),
  KEY "id" ("id"),
  KEY "address" ("address")
);

CREATE TABLE "property" (
  "id" varchar(24) NOT NULL,
  "collection_id" varchar(24) NOT NULL,
  "name" varchar(128) NOT NULL,
  PRIMARY KEY ("id","name")
);

CREATE TABLE "property_value" (
  "id" varchar(24) NOT NULL,
  "name" varchar(128) NOT NULL,
  "image" varchar(256) DEFAULT NULL,
  "image_integrity" varchar(256) DEFAULT NULL,
  "image_mimetype" varchar(32) DEFAULT NULL,
  "animation_url" varchar(256) DEFAULT NULL,
  "animation_url_integrity" varchar(256) DEFAULT NULL,
  "animation_url_mimetype" varchar(32) DEFAULT NULL,
  PRIMARY KEY ("id","name"),
  UNIQUE KEY "col_property_name" ("id","name"),
  KEY "image" ("image")
);

CREATE TABLE "property_value_extras" (
  "id" varchar(24) NOT NULL,
  "name" varchar(128) NOT NULL,
  "mkey" varchar(128) NOT NULL,
  "mvalue" text NOT NULL,
  PRIMARY KEY ("id","name","mkey"),
  KEY "name" ("name"),
  KEY "key" ("mkey")
);

CREATE TABLE "collection_extras" (
  "id" varchar(24) NOT NULL,
  "mkey" varchar(128) NOT NULL,
  "mvalue" text NOT NULL,
  PRIMARY KEY ("id","mkey"),
  KEY "key" ("mkey")
);

CREATE TABLE "community_faq" (
  "id" bigint unsigned NOT NULL,
  "q" varchar(256) NOT NULL,
  "a" text NOT NULL,
  "ordering" int unsigned NOT NULL,
  PRIMARY KEY ("id","q")
);

CREATE TABLE "community_extras" (
  "id" bigint unsigned NOT NULL,
  "mkey" varchar(128) NOT NULL,
  "mvalue" text NOT NULL,
  PRIMARY KEY ("id","mkey"),
  KEY "key" ("mkey")
);