export type MemberRole =
  | "admin"
  | "staff"
  | "auditor"
  | "member"
  | "agent_member"
  | "master_member"
  | "user"
  | "agent"
  | "master"
  | "marketing"
  | "analis"
  | "operator_trx"
  | "operator_wallet";
export type ManageableRole =
  | "staff"
  | "auditor"
  | "member"
  | "agent_member"
  | "master_member"
  | "user"
  | "agent"
  | "master"
  | "marketing"
  | "analis"
  | "operator_trx"
  | "operator_wallet";

export type AccountScope = "h2h" | "retail" | "internal";

export const ALL_MEMBER_ROLES: MemberRole[] = [
  "admin",
  "staff",
  "auditor",
  "member",
  "agent_member",
  "master_member",
  "user",
  "agent",
  "master",
  "marketing",
  "analis",
  "operator_trx",
  "operator_wallet",
];

export const CREATE_USER_ROLES: ManageableRole[] = [
  "staff",
  "auditor",
  "member",
  "agent_member",
  "master_member",
  "user",
  "agent",
  "master",
  "marketing",
  "analis",
  "operator_trx",
  "operator_wallet",
];

export const H2H_ROLES: ManageableRole[] = ["member", "agent_member", "master_member"];
export const RETAIL_ROLES: ManageableRole[] = ["user", "agent", "master", "marketing", "analis"];
export const INTERNAL_ROLES: ManageableRole[] = ["staff", "auditor", "operator_trx", "operator_wallet"];
export const INTERNAL_ALL_ROLES: MemberRole[] = ["admin", "staff", "auditor", "operator_trx", "operator_wallet"];

export const ACCOUNT_SCOPE_OPTIONS: Array<{ value: AccountScope; label: string; description: string }> = [
  { value: "h2h", label: "Member H2H", description: "Akun API key, fee kategori, dan transaksi H2H." },
  { value: "retail", label: "Member Retail", description: "Akun user, agent, dan master retail." },
  { value: "internal", label: "Akun Internal", description: "Admin dan operator internal." },
];

export const ROLE_OPTIONS = ALL_MEMBER_ROLES.map((value) => ({
  value,
  label: roleLabel(value),
}));

export const CREATE_ROLE_OPTIONS = CREATE_USER_ROLES.map((value) => ({
  value,
  label: roleLabel(value),
}));

export function rolesForScope(scope: AccountScope): MemberRole[] {
  switch (scope) {
    case "h2h":
      return H2H_ROLES;
    case "retail":
      return RETAIL_ROLES;
    case "internal":
      return INTERNAL_ALL_ROLES;
    default:
      return ALL_MEMBER_ROLES;
  }
}

export function createRolesForScope(scope: AccountScope): ManageableRole[] {
  switch (scope) {
    case "h2h":
      return H2H_ROLES;
    case "retail":
      return RETAIL_ROLES;
    case "internal":
      return INTERNAL_ROLES;
    default:
      return CREATE_USER_ROLES;
  }
}

export function roleLabel(role: string): string {
  const r = String(role || "").trim().toLowerCase();
  switch (r) {
    case "admin":
      return "Admin";
    case "staff":
      return "Staff";
    case "auditor":
      return "Auditor";
    case "member":
      return "Member";
    case "agent_member":
      return "Agent Member";
    case "master_member":
      return "Master Member";
    case "user":
      return "User";
    case "agent":
      return "Agent";
    case "master":
      return "Master";
    case "marketing":
      return "Marketing";
    case "analis":
      return "Operator";
    case "operator_trx":
      return "Operator Transaksi";
    case "operator_wallet":
      return "Operator Wallet";
    default:
      return "Semua Role";
  }
}
