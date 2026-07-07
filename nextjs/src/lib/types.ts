export interface TempleImage {
  id: number;
  temple_id: number;
  url: string;
  alt: string;
  sort_order: number;
}

export interface Temple {
  id: number;
  name: string;
  slug: string;
  state: string;
  city: string;
  deity: string;
  arch_style: string;
  description: string;
  history: string;
  image_url: string;
  model_url: string;
  ambient_audio: string;
  latitude: number;
  longitude: number;
  visit_duration: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
  images?: TempleImage[];
}

export interface User {
  id: number;
  email: string;
  name: string;
  role: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Favorite {
  id: number;
  user_id: number;
  temple_id: number;
  temple_slug: string;
  temple_name: string;
  created_at: string;
}

export interface Review {
  id: number;
  user_id: number;
  temple_id: number;
  rating: number;
  text: string;
  user_name: string;
  created_at: string;
  updated_at: string;
}

export interface PopularTemple {
  id: number;
  name: string;
  slug: string;
  state: string;
  views: number;
}
