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

INSERT INTO public.gateway (gtw_adapter_id, gtw_params_jsonb)
VALUES ('paylink', '{"transport": {"base_address": "https://payadap.co/api/v1/"}, "payment_methods": [{"id": "p2pcard", "name": "P2P Card", "required_fields": []}]}'),
 ('auris', '{"transport": {"base_address": "https://changer.icu/api2"}, "payment_methods": [{"id": "p2pcard", "gtw_id": {"KGS": 237, "KZT": 260, "RUB": 86}, "required_fields": []}, {"id": "p2psbp", "gtw_id": {"RUB": 88}, "required_fields": []}]}'),
 ('sequoia', '{"transport": {"base_address": "https://sequoiapay.com"}, "payment_methods": [{"id": "p2pcard", "name": "P2P Card", "required_fields": [{"id": "bank", "name": "Bank", "type": "select", "values": [{"id": "ru_sberbank", "name": "Sberbank"}]}]}]}'),
 ('asupayme', '{"transport": {"base_address":"http://147.45.152.131:8021"}}');


INSERT INTO public.channel (chn_name, gtw_id, chn_params_jsonb)
VALUES ('paylink', 1, '{"credentials": {"api_key": "m9bb1lelXNK3c198UWb2e41J1EQM", "merch_id": "52161-b615-4211-8638-3998bdb806c0"}, "payment_methods": ["p2pcard"]}'),
 ('auris', 2, '{"credentials": {"api_key": "049050051052-8ECG0WKvwSTWBzHTv2qtzy6MTiCkEoT2QAVAmNoGfQPzm02sPuBtz4nRwCJRs6fmM0WHUNqke6gG8o1MKcVHZGpJScZtOt8SmC3X1PGRx9uMTL3rw-OQP244B0M3HTID6T8PixLROqwaYQHJq6", "shop_id": 1102, "secret_key": "049050051052-jVDNvl8Lm8U4LQ7NPFGP2K03-k7w0o06k"}, "payment_methods": ["p2pcard"]}'),
 ('sequoia', 3, '{"credentials": {"secret_key": "8f5DSbXOolXwhLVmtOd", "callback_secret": "fqLFXw9BQLbW7G83Uf"}, "payment_methods": ["p2pcard"]}'),
 ('asupayme', 4, e'{ "credentials": {"secret_key": "superSecretKey228", "api_key": "c2RmZHNhYnZmZGFiYWV0dnJ0c3JhZGZoYnN0cmRmdmJzZ2Z4MjM0NTQzZ2czcXZhZQ==", "merchant_id": "32"}}');


