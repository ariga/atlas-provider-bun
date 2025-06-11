-- Create "stories" table
CREATE TABLE [stories] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [title] varchar(255) NULL,
  [author_id] bigint NULL,
  CONSTRAINT [PK__stories__3213E83F94D4B7DF] PRIMARY KEY CLUSTERED ([id] ASC)
);
-- Create "users" table
CREATE TABLE [users] (
  [id] bigint IDENTITY (1, 1) NOT NULL,
  [name] varchar(255) NULL,
  [emails] varchar(255) NULL,
  CONSTRAINT [PK__users__3213E83F57564211] PRIMARY KEY CLUSTERED ([id] ASC)
);
