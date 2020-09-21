TRUNCATE users CASCADE; -- delete everything in cascade

INSERT INTO users (id, login, hashed_password)
VALUES ('bbc00191-b064-4655-9075-261ccef978cb', 'user_a', 'user_a_pass'),
       ('f65697a1-dbe7-49b5-93d6-bbfc512a46f6', 'user_b', 'user_b_pass'),
       ('88354f85-d784-467f-b5dc-5260d173853f', 'user_c', 'user_c_pass'),
       ('6aacb72a-264d-4bc3-b2f9-9fb26a78a449', 'user_d', 'user_d_pass');

INSERT INTO wallets (id, user_id, balance)
VALUES ('2f9b76dd-f689-456e-9080-6789718018a5', 'bbc00191-b064-4655-9075-261ccef978cb', 12.75),
       ('4e1d841d-e53f-4785-ba4d-99df05f11eee', 'f65697a1-dbe7-49b5-93d6-bbfc512a46f6', 52.25),
       ('f0212317-88db-4dd4-ba0e-39757e1ebcc6', '88354f85-d784-467f-b5dc-5260d173853f',  0.00),
       ('f889299f-41c4-4e58-96c2-7451c8276842', '6aacb72a-264d-4bc3-b2f9-9fb26a78a449', 30.50);

INSERT INTO transfers (id, issuer_id, origin_wallet_id, destination_wallet_id, amount, date, message)
VALUES ('97ca2b73-7988-4247-82d4-f6ba723a99c9', 'bbc00191-b064-4655-9075-261ccef978cb', '2f9b76dd-f689-456e-9080-6789718018a5', '4e1d841d-e53f-4785-ba4d-99df05f11eee', 7.25, '2020-09-19 10:00:02+00:00', 'dinner');

INSERT INTO transactions (id, wallet_id, amount, balance, date, transaction_type, reference_id)
VALUES ('9177ad78-e5d5-4d3c-be8c-e0e1f44bbdcc', '2f9b76dd-f689-456e-9080-6789718018a5', 20.00, 20.00, '2020-09-20 10:00:00+00:00', 'deposit', NULL),
       ('7b63d966-a700-4d4f-b12c-e490bc96fd8c', '4e1d841d-e53f-4785-ba4d-99df05f11eee', 45.00, 45.00, '2020-09-20 10:00:01+00:00', 'deposit', NULL),
       ('72db21de-9a63-40c6-b666-d35cf8437fd5', 'f889299f-41c4-4e58-96c2-7451c8276842',  9.50,  9.50, '2020-09-20 10:00:01+00:00', 'deposit', NULL),
       ('be3c602a-4410-475b-b39f-016c451726a1', 'f889299f-41c4-4e58-96c2-7451c8276842', 08.50, 18.00, '2020-09-20 10:00:02+00:00', 'deposit', NULL),
       ('ccf28188-92c7-4a30-8f60-70345694f893', 'f889299f-41c4-4e58-96c2-7451c8276842', 12.50, 30.50, '2020-09-20 10:00:02+00:00', 'deposit', NULL),
       ('4bce4401-6b35-4fa1-94b9-ac5ce05d29b1', '2f9b76dd-f689-456e-9080-6789718018a5', -7.25, 12.75, '2020-09-20 11:10:00+00:00', 'transfer', '97ca2b73-7988-4247-82d4-f6ba723a99c9'),
       ('60978032-b118-4727-a123-4468dced4104', '4e1d841d-e53f-4785-ba4d-99df05f11eee',  7.25, 52.25, '2020-09-20 11:10:00+00:00', 'transfer', '97ca2b73-7988-4247-82d4-f6ba723a99c9');
