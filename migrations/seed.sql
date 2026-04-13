-- Seed data for local development / testing
-- User credentials: test@example.com / password123

INSERT INTO users (id, name, email, password_hash)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Test User',
    'test@example.com',
    '$2a$10$BvpUs5JJ9EP0kvZ8lcEk8eYsXhNoN2iKW6A2BTTzXqQXRtKr1sSym'
) ON CONFLICT DO NOTHING;

INSERT INTO projects (id, name, description, owner_id)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'Demo Project',
    'A sample project for local development',
    'a0000000-0000-0000-0000-000000000001'
) ON CONFLICT DO NOTHING;

INSERT INTO tasks (id, project_id, title, status, priority, assignee_id)
VALUES
    (
        'c0000000-0000-0000-0000-000000000001',
        'b0000000-0000-0000-0000-000000000001',
        'Set up repository',
        'done',
        'high',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'c0000000-0000-0000-0000-000000000002',
        'b0000000-0000-0000-0000-000000000001',
        'Implement authentication',
        'in_progress',
        'high',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'c0000000-0000-0000-0000-000000000003',
        'b0000000-0000-0000-0000-000000000001',
        'Write API documentation',
        'todo',
        'medium',
        NULL
    )
ON CONFLICT DO NOTHING;
