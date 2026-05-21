-- =============================================================
-- VENUE SUBMISSION SEED
-- Inserts pending venue submissions for real Stockholm bars.
-- Accept them via POST /admin/submission/:id/accept to trigger
-- the full flow (venue creation + business hours via get-places-data).
--
-- Requires at least one existing user in auth.users.
-- Run after the main seed.sql.
-- =============================================================

do $$
declare
    submitter uuid;
begin
    select id into submitter from auth.users limit 1;
    if submitter is null then
        raise exception 'No users found in auth.users. Create a user account first.';
    end if;

    insert into public.submission (submitted_by, category, status, payload) values

      -- Kvarnen — iconic Stockholm beer hall since 1908, Södermalm
      ( submitter, 'venue', 'pending',
        '{"name":"Kvarnen","street":"Tjärhovsgatan 4","area":"Södermalm","city":"Stockholm","country":"Sweden","zip":"116 21","lat":59.316617,"lng":18.074761}'
      ),

      -- Akkurat — cellar bar known for Belgian beers and whisky, Södermalm
      ( submitter, 'venue', 'pending',
        '{"name":"Akkurat","street":"Hornsgatan 18","area":"Södermalm","city":"Stockholm","country":"Sweden","zip":"118 20","lat":59.318127,"lng":18.064700}'
      ),

      -- Gröne Jäger — traditional Swedish pub, Vasastan
      ( submitter, 'venue', 'pending',
        '{"name":"Gröne Jäger","street":"Surbrunnsgatan 16","area":"Vasastan","city":"Stockholm","country":"Sweden","zip":"113 48","lat":59.343221,"lng":18.055599}'
      );

end $$;
