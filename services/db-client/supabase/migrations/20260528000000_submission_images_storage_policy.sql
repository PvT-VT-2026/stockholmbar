-- Allow authenticated users to upload to the submission-images bucket.
-- This is needed so app users can submit menu photos.
CREATE POLICY "Authenticated users can upload menu images"
ON storage.objects
FOR INSERT
TO authenticated
WITH CHECK (bucket_id = 'submission-images');

-- Allow public read access so the stored URLs work in the admin review screen.
CREATE POLICY "Public read access for menu images"
ON storage.objects
FOR SELECT
TO public
USING (bucket_id = 'submission-images');
