CREATE OR REPLACE FUNCTION "public"."handle_new_user"() RETURNS "trigger"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO ''
    AS $$
begin
  insert into public.profile (id, email, display_name)
  values (
        new.id,
        new.email,
        new.raw_user_meta_data ->> 'display_name'
    );
  return new;
end;
$$;
