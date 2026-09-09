create table if not exists public.enterprise_orgs (
  id uuid primary key default gen_random_uuid(),
  domain text unique not null,
  saml_provider_id text,
  plan text default 'enterprise',
  seats_purchased int default 10,
  created_at timestamptz default now()
);

-- Add org_id to users table
alter table public.users add column if not exists org_id uuid references public.enterprise_orgs(id);

-- Create trigger to automatically grant enterprise plan and link org_id on SSO login
create or replace function public.handle_sso_login()
returns trigger as $$
declare
  _domain text;
  _org_id uuid;
begin
  -- Check if the user is logging in via SSO (SAML)
  if new.raw_app_meta_data->>'provider' = 'sso' then
    -- Extract the domain from the user's email
    _domain := split_part(new.email, '@', 2);

    -- Find the corresponding organization by domain
    select id into _org_id from public.enterprise_orgs where domain = _domain;

    if _org_id is not null then
      -- Link the user to the organization and set their plan to 'enterprise'
      update public.users
      set org_id = _org_id, plan = 'enterprise'
      where id = new.id;
    end if;
  end if;

  return new;
end;
$$ language plpgsql security definer;

-- Drop trigger if exists to allow safe rerunning
drop trigger if exists on_auth_user_sso_login on auth.users;

-- Create the trigger on the auth.users table
create trigger on_auth_user_sso_login
  after insert or update on auth.users
  for each row
  execute procedure public.handle_sso_login();
