export interface Guest {
  id: number;
  first_name: string;
  last_name: string;
  phone: string | null;
  relationship: string;
  confirmed: boolean;
  family_group: number;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

export interface CreateGuestInput {
  first_name: string;
  last_name: string;
  phone: string;
  relationship: string;
  family_group: number;
}

export interface UpdateGuestInput {
  first_name?: string;
  last_name?: string;
  phone?: string;
  relationship?: string;
  confirmed?: boolean;
  family_group?: number;
}

export interface ImportResult {
  imported: number;
  errors: string[];
  total: number;
}
