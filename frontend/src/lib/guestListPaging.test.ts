import { describe, test, expect } from "bun:test";
import { nextPageOnSearchChange } from "./guestListPaging";

describe("nextPageOnSearchChange — estado da paginação na busca", () => {
  test("ao iniciar busca, vai para página 1 e guarda a página atual", () => {
    const result = nextPageOnSearchChange("", "maria", 2, null);
    expect(result.page).toBe(1);
    expect(result.savedPage).toBe(2);
  });

  test("regressão (bug 3): ao limpar a busca, restaura a página em que estava", () => {
    const result = nextPageOnSearchChange("maria", "", 1, 2);
    expect(result.page).toBe(2);
    expect(result.savedPage).toBeNull();
  });

  test("ao trocar o termo de busca, volta para página 1", () => {
    const result = nextPageOnSearchChange("maria", "joão", 1, 3);
    expect(result.page).toBe(1);
  });

  test("sem mudança real no termo, mantém a página atual", () => {
    const result = nextPageOnSearchChange("maria", "maria", 4, 2);
    expect(result.page).toBe(4);
    expect(result.savedPage).toBe(2);
  });

  test("limpar sem página salva cai na página 1", () => {
    const result = nextPageOnSearchChange("maria", "", 1, null);
    expect(result.page).toBe(1);
  });
});
