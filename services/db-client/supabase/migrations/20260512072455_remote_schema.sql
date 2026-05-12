


SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


COMMENT ON SCHEMA "public" IS 'standard public schema';



CREATE EXTENSION IF NOT EXISTS "pg_stat_statements" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "supabase_vault" WITH SCHEMA "vault";






CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA "extensions";






CREATE OR REPLACE FUNCTION "public"."handle_new_user"() RETURNS "trigger"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO ''
    AS $$
begin
  insert into public.profiles (id, email, display_name)
  values (
        new.id,
        new.email,
        new.raw_user_meta_data ->> 'display_name'
    );
  return new;
end;
$$;


ALTER FUNCTION "public"."handle_new_user"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."is_admin"() RETURNS boolean
    LANGUAGE "sql" STABLE
    AS $$
  select exists (
    select 1
    from public."profile"
    where id = auth.uid()
      and role = 'admin'
  );
$$;


ALTER FUNCTION "public"."is_admin"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."set_updated_at"() RETURNS "trigger"
    LANGUAGE "plpgsql"
    AS $$
begin
  new.updated_at = now();
  return new;
end;
$$;


ALTER FUNCTION "public"."set_updated_at"() OWNER TO "postgres";

SET default_tablespace = '';

SET default_table_access_method = "heap";


CREATE TABLE IF NOT EXISTS "public"."beverage" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "name" "text" NOT NULL,
    "abv" numeric(5,2),
    "description" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone
);


ALTER TABLE "public"."beverage" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."business_hours" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "venue_id" "uuid" NOT NULL,
    "day_of_week" smallint NOT NULL,
    "open_time" time without time zone,
    "close_time" time without time zone,
    "is_closed" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone,
    CONSTRAINT "business_hours_day_of_week_check" CHECK ((("day_of_week" >= 0) AND ("day_of_week" <= 6)))
);


ALTER TABLE "public"."business_hours" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."happy_hours" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "venue_id" "uuid" NOT NULL,
    "day_of_week" smallint NOT NULL,
    "start_time" time without time zone NOT NULL,
    "end_time" time without time zone NOT NULL,
    "is_active" boolean DEFAULT true NOT NULL,
    "name" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone,
    CONSTRAINT "happy_hours_day_of_week_check" CHECK ((("day_of_week" >= 0) AND ("day_of_week" <= 6)))
);


ALTER TABLE "public"."happy_hours" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."location" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "street" "text",
    "area" "text",
    "city" "text",
    "country" "text",
    "zip" "text",
    "lat" numeric(9,6),
    "lng" numeric(9,6),
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone
);


ALTER TABLE "public"."location" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."price_record" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "venue_unit_id" "uuid" NOT NULL,
    "currency" "text" NOT NULL,
    "amount" numeric(10,2) NOT NULL,
    "recorded_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone
);


ALTER TABLE "public"."price_record" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."profile" (
    "id" "uuid" NOT NULL,
    "email" "text",
    "display_name" "text",
    "avatar_url" "text",
    "role" "text" DEFAULT 'user'::"text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "profile_role_check" CHECK (("role" = ANY (ARRAY['admin'::"text", 'user'::"text"])))
);


ALTER TABLE "public"."profile" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."submission" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "submitted_by" "uuid" NOT NULL,
    "category" "text" NOT NULL,
    "status" "text" NOT NULL,
    "payload" "jsonb" NOT NULL,
    "reviewed_at" timestamp with time zone,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone,
    "payload_hash" "text",
    CONSTRAINT "submission_status_check" CHECK (("status" = ANY (ARRAY['pending'::"text", 'accepted'::"text", 'rejected'::"text"])))
);


ALTER TABLE "public"."submission" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."submission_image" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "submission_id" "uuid" NOT NULL,
    "data" "bytea" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."submission_image" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."unit" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "beverage_id" "uuid" NOT NULL,
    "name" "text" NOT NULL,
    "volume_ml" integer,
    "unit_type" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone,
    "size" "text"
);


ALTER TABLE "public"."unit" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."venue" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "location_id" "uuid" NOT NULL,
    "venue_chain_id" "uuid",
    "name" "text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone
);


ALTER TABLE "public"."venue" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."venue_chain" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "name" "text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone
);


ALTER TABLE "public"."venue_chain" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."venue_unit" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "venue_id" "uuid" NOT NULL,
    "unit_id" "uuid" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "deleted_at" timestamp with time zone
);


ALTER TABLE "public"."venue_unit" OWNER TO "postgres";


ALTER TABLE ONLY "public"."beverage"
    ADD CONSTRAINT "beverage_name_abv_unique" UNIQUE ("name", "abv");



ALTER TABLE ONLY "public"."beverage"
    ADD CONSTRAINT "beverage_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."business_hours"
    ADD CONSTRAINT "business_hours_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."business_hours"
    ADD CONSTRAINT "business_hours_venue_day_unique" UNIQUE ("venue_id", "day_of_week");



ALTER TABLE ONLY "public"."happy_hours"
    ADD CONSTRAINT "happy_hours_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."happy_hours"
    ADD CONSTRAINT "happy_hours_venue_day_start_unique" UNIQUE ("venue_id", "day_of_week", "start_time");



ALTER TABLE ONLY "public"."location"
    ADD CONSTRAINT "location_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."price_record"
    ADD CONSTRAINT "price_record_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."profile"
    ADD CONSTRAINT "profile_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."submission_image"
    ADD CONSTRAINT "submission_image_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."submission_image"
    ADD CONSTRAINT "submission_image_submission_id_key" UNIQUE ("submission_id");



ALTER TABLE ONLY "public"."submission"
    ADD CONSTRAINT "submission_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."unit"
    ADD CONSTRAINT "unit_beverage_volume_type_unique" UNIQUE ("beverage_id", "volume_ml", "unit_type");



ALTER TABLE ONLY "public"."unit"
    ADD CONSTRAINT "unit_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."venue_chain"
    ADD CONSTRAINT "venue_chain_name_unique" UNIQUE ("name");



ALTER TABLE ONLY "public"."venue_chain"
    ADD CONSTRAINT "venue_chain_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."venue"
    ADD CONSTRAINT "venue_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."venue_unit"
    ADD CONSTRAINT "venue_unit_pkey" PRIMARY KEY ("id");



CREATE INDEX "beverage_deleted_at_idx" ON "public"."beverage" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "business_hours_deleted_at_idx" ON "public"."business_hours" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "business_hours_venue_id_idx" ON "public"."business_hours" USING "btree" ("venue_id");



CREATE INDEX "happy_hours_deleted_at_idx" ON "public"."happy_hours" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "happy_hours_venue_id_idx" ON "public"."happy_hours" USING "btree" ("venue_id");



CREATE INDEX "price_record_deleted_at_idx" ON "public"."price_record" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "price_record_venue_unit_id_idx" ON "public"."price_record" USING "btree" ("venue_unit_id");



CREATE UNIQUE INDEX "submission_user_payload_unique" ON "public"."submission" USING "btree" ("submitted_by", "payload_hash");



CREATE INDEX "unit_beverage_id_idx" ON "public"."unit" USING "btree" ("beverage_id");



CREATE INDEX "unit_deleted_at_idx" ON "public"."unit" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "venue_deleted_at_idx" ON "public"."venue" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "venue_location_id_idx" ON "public"."venue" USING "btree" ("location_id");



CREATE INDEX "venue_unit_deleted_at_idx" ON "public"."venue_unit" USING "btree" ("deleted_at") WHERE ("deleted_at" IS NULL);



CREATE INDEX "venue_unit_unit_id_idx" ON "public"."venue_unit" USING "btree" ("unit_id");



CREATE INDEX "venue_unit_venue_id_idx" ON "public"."venue_unit" USING "btree" ("venue_id");



CREATE UNIQUE INDEX "venue_unit_venue_id_unit_id_idx" ON "public"."venue_unit" USING "btree" ("venue_id", "unit_id") WHERE ("deleted_at" IS NULL);



CREATE INDEX "venue_venue_chain_id_idx" ON "public"."venue" USING "btree" ("venue_chain_id");



CREATE OR REPLACE TRIGGER "trg_beverage_updated_at" BEFORE UPDATE ON "public"."beverage" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_business_hours_updated_at" BEFORE UPDATE ON "public"."business_hours" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_happy_hours_updated_at" BEFORE UPDATE ON "public"."happy_hours" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_location_updated_at" BEFORE UPDATE ON "public"."location" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_price_record_updated_at" BEFORE UPDATE ON "public"."price_record" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_unit_updated_at" BEFORE UPDATE ON "public"."unit" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_venue_chain_updated_at" BEFORE UPDATE ON "public"."venue_chain" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_venue_unit_updated_at" BEFORE UPDATE ON "public"."venue_unit" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



CREATE OR REPLACE TRIGGER "trg_venue_updated_at" BEFORE UPDATE ON "public"."venue" FOR EACH ROW EXECUTE FUNCTION "public"."set_updated_at"();



ALTER TABLE ONLY "public"."business_hours"
    ADD CONSTRAINT "business_hours_venue_id_fkey" FOREIGN KEY ("venue_id") REFERENCES "public"."venue"("id");



ALTER TABLE ONLY "public"."happy_hours"
    ADD CONSTRAINT "happy_hours_venue_id_fkey" FOREIGN KEY ("venue_id") REFERENCES "public"."venue"("id");



ALTER TABLE ONLY "public"."price_record"
    ADD CONSTRAINT "price_record_venue_unit_id_fkey" FOREIGN KEY ("venue_unit_id") REFERENCES "public"."venue_unit"("id");



ALTER TABLE ONLY "public"."profile"
    ADD CONSTRAINT "profile_id_fkey" FOREIGN KEY ("id") REFERENCES "auth"."users"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."submission_image"
    ADD CONSTRAINT "submission_image_submission_id_fkey" FOREIGN KEY ("submission_id") REFERENCES "public"."submission"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."submission"
    ADD CONSTRAINT "submission_submitted_by_fkey" FOREIGN KEY ("submitted_by") REFERENCES "auth"."users"("id");



ALTER TABLE ONLY "public"."unit"
    ADD CONSTRAINT "unit_beverage_id_fkey" FOREIGN KEY ("beverage_id") REFERENCES "public"."beverage"("id");



ALTER TABLE ONLY "public"."venue"
    ADD CONSTRAINT "venue_location_id_fkey" FOREIGN KEY ("location_id") REFERENCES "public"."location"("id");



ALTER TABLE ONLY "public"."venue_unit"
    ADD CONSTRAINT "venue_unit_unit_id_fkey" FOREIGN KEY ("unit_id") REFERENCES "public"."unit"("id");



ALTER TABLE ONLY "public"."venue_unit"
    ADD CONSTRAINT "venue_unit_venue_id_fkey" FOREIGN KEY ("venue_id") REFERENCES "public"."venue"("id");



ALTER TABLE ONLY "public"."venue"
    ADD CONSTRAINT "venue_venue_chain_id_fkey" FOREIGN KEY ("venue_chain_id") REFERENCES "public"."venue_chain"("id");



CREATE POLICY "Admin manage beverage" ON "public"."beverage" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage business_hours" ON "public"."business_hours" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage happy_hours" ON "public"."happy_hours" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage location" ON "public"."location" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage price_record" ON "public"."price_record" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage unit" ON "public"."unit" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage venue_chain" ON "public"."venue_chain" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage venue_unit" ON "public"."venue_unit" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admin manage venues" ON "public"."venue" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Admins full access" ON "public"."profile" USING ("public"."is_admin"()) WITH CHECK ("public"."is_admin"());



CREATE POLICY "Public read beverage" ON "public"."beverage" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read business_hours" ON "public"."business_hours" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read happy_hours" ON "public"."happy_hours" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read location" ON "public"."location" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read price_record" ON "public"."price_record" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read unit" ON "public"."unit" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read venue_chain" ON "public"."venue_chain" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read venue_unit" ON "public"."venue_unit" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Public read venues" ON "public"."venue" FOR SELECT USING (("deleted_at" IS NULL));



CREATE POLICY "Users can update own profile" ON "public"."profile" FOR UPDATE USING (("id" = "auth"."uid"())) WITH CHECK (("id" = "auth"."uid"()));



CREATE POLICY "Users can view own profile" ON "public"."profile" FOR SELECT USING (("id" = "auth"."uid"()));



ALTER TABLE "public"."beverage" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."business_hours" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."happy_hours" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."location" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."price_record" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."profile" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."submission" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."submission_image" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."unit" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."venue" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."venue_chain" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."venue_unit" ENABLE ROW LEVEL SECURITY;




ALTER PUBLICATION "supabase_realtime" OWNER TO "postgres";


GRANT USAGE ON SCHEMA "public" TO "postgres";
GRANT USAGE ON SCHEMA "public" TO "anon";
GRANT USAGE ON SCHEMA "public" TO "authenticated";
GRANT USAGE ON SCHEMA "public" TO "service_role";






















































































































































GRANT ALL ON FUNCTION "public"."handle_new_user"() TO "anon";
GRANT ALL ON FUNCTION "public"."handle_new_user"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."handle_new_user"() TO "service_role";



GRANT ALL ON FUNCTION "public"."is_admin"() TO "anon";
GRANT ALL ON FUNCTION "public"."is_admin"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."is_admin"() TO "service_role";



GRANT ALL ON FUNCTION "public"."set_updated_at"() TO "anon";
GRANT ALL ON FUNCTION "public"."set_updated_at"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."set_updated_at"() TO "service_role";


















GRANT ALL ON TABLE "public"."beverage" TO "anon";
GRANT ALL ON TABLE "public"."beverage" TO "authenticated";
GRANT ALL ON TABLE "public"."beverage" TO "service_role";



GRANT ALL ON TABLE "public"."business_hours" TO "anon";
GRANT ALL ON TABLE "public"."business_hours" TO "authenticated";
GRANT ALL ON TABLE "public"."business_hours" TO "service_role";



GRANT ALL ON TABLE "public"."happy_hours" TO "anon";
GRANT ALL ON TABLE "public"."happy_hours" TO "authenticated";
GRANT ALL ON TABLE "public"."happy_hours" TO "service_role";



GRANT ALL ON TABLE "public"."location" TO "anon";
GRANT ALL ON TABLE "public"."location" TO "authenticated";
GRANT ALL ON TABLE "public"."location" TO "service_role";



GRANT ALL ON TABLE "public"."price_record" TO "anon";
GRANT ALL ON TABLE "public"."price_record" TO "authenticated";
GRANT ALL ON TABLE "public"."price_record" TO "service_role";



GRANT ALL ON TABLE "public"."profile" TO "anon";
GRANT ALL ON TABLE "public"."profile" TO "authenticated";
GRANT ALL ON TABLE "public"."profile" TO "service_role";



GRANT ALL ON TABLE "public"."submission" TO "anon";
GRANT ALL ON TABLE "public"."submission" TO "authenticated";
GRANT ALL ON TABLE "public"."submission" TO "service_role";



GRANT ALL ON TABLE "public"."submission_image" TO "anon";
GRANT ALL ON TABLE "public"."submission_image" TO "authenticated";
GRANT ALL ON TABLE "public"."submission_image" TO "service_role";



GRANT ALL ON TABLE "public"."unit" TO "anon";
GRANT ALL ON TABLE "public"."unit" TO "authenticated";
GRANT ALL ON TABLE "public"."unit" TO "service_role";



GRANT ALL ON TABLE "public"."venue" TO "anon";
GRANT ALL ON TABLE "public"."venue" TO "authenticated";
GRANT ALL ON TABLE "public"."venue" TO "service_role";



GRANT ALL ON TABLE "public"."venue_chain" TO "anon";
GRANT ALL ON TABLE "public"."venue_chain" TO "authenticated";
GRANT ALL ON TABLE "public"."venue_chain" TO "service_role";



GRANT ALL ON TABLE "public"."venue_unit" TO "anon";
GRANT ALL ON TABLE "public"."venue_unit" TO "authenticated";
GRANT ALL ON TABLE "public"."venue_unit" TO "service_role";









ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "service_role";































drop extension if exists "pg_net";

CREATE TRIGGER on_auth_user_created AFTER INSERT ON auth.users FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();


