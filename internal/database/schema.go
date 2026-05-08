package database

var schema string = `CREATE TABLE IF NOT EXISTS nodes (
id TEXT PRIMARY KEY,
    parent_id TEXT,               -- Points to the parent node's ID (NULL if root)
    batch_id TEXT NOT NULL,       -- For our atomic undo functionality
    sequence INTEGER NOT NULL,    -- To maintain order of notes within the same parent
    node_type TEXT NOT NULL,      -- e.g., 'subject', 'book', 'chapter', 'topic', 'note'
    label TEXT,                   -- The name of the meta tag (e.g., "Advanced Go", "Chapter 1")
    content TEXT,                 -- The actual text (Usually only populated if node_type == 'note')
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY(parent_id) REFERENCES nodes(id) ON DELETE CASCADE
);`

var index_creation string = `
-- Indexes for fast tree traversal and undo operations
CREATE INDEX IF NOT EXISTS idx_nodes_parent ON nodes(parent_id);
CREATE INDEX IF NOT EXISTS idx_nodes_batch ON nodes(batch_id);
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(node_type);`
