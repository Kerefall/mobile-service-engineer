/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    dangerouslyAllowSVG: true,
    remotePatterns: [], // Base64 не требует remotePatterns
  },
};
export default nextConfig;