INSERT INTO users(id, first_name, last_name, nickname, email, country) VALUES
('00000000-0000-0000-0000-000000000001', 'John', 'Doe', 'johndoe', 'johndoe@email.com', 'US'),
('00000000-0000-0000-0000-000000000002', 'Jane', 'Doe', 'janedoe', 'janedoe@email.com', 'UK'),
('00000000-0000-0000-0000-000000000003', 'Zoltan', 'Domahidi', 'dome', 'dome@email.com', 'UK')
ON CONFLICT DO NOTHING;
