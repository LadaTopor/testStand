create table gateway
(
    gtw_id           integer generated always as identity
        primary key,
    gtw_is_active    boolean default true not null,
    gtw_adapter_id   varchar(30)          not null,
    gtw_params_jsonb jsonb
);

create table channel
(
    chn_id           integer generated always as identity
        primary key,
    chn_name         varchar              not null,
    chn_is_active    boolean default true not null,
    gtw_id           integer              not null
        references gateway,
    chn_params_jsonb jsonb
);