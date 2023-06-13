import { useCallback } from "react";
import type { Organization, Scenario } from "@/models";
import {
  type OrganizationRepository,
  type ScenariosRepository,
  type UserRepository,
  fetchAllOrganizations,
  fetchOrganization,
  postOrganization,
  fetchScenarios,
  deleteOrganization,
} from "@/repositories";
import { useSimpleLoader } from "@/hooks/SimpleLoader";
import { type LoadingDispatcher } from "@/hooks/Loading";

export interface OrganizationService {
  organizationRepository: OrganizationRepository;
  scenariosRepository: ScenariosRepository;
  userRepository: UserRepository;
}

export function useAllOrganizations(
  service: OrganizationService,
  loadingDispatcher: LoadingDispatcher
) {
  const loadAllOrganizations = useCallback(() => {
    return fetchAllOrganizations(service.organizationRepository);
  }, [service]);

  const [allOrganizations, refreshAllOrganizations] = useSimpleLoader<
    Organization[]
  >(loadingDispatcher, loadAllOrganizations);

  return {
    allOrganizations,
    refreshAllOrganizations,
  };
}

export function useOrganization(
  service: OrganizationService,
  loadingDispatcher: LoadingDispatcher,
  organizationId: string
) {
  const loadOrganization = useCallback(() => {
    return fetchOrganization(service.organizationRepository, organizationId);
  }, [service, organizationId]);

  const [organization, refreshOrganization] = useSimpleLoader<Organization>(
    loadingDispatcher,
    loadOrganization
  );

  return {
    organization,
    refreshOrganization,
  };
}

export function useCreateOrganization(service: OrganizationService) {
  const createOrganization = useCallback(
    async (name: string) => {
      await postOrganization(service.organizationRepository, {
        name,
        databaseName: name,
      });
    },
    [service]
  );

  return {
    createOrganization,
  };
}

export function useDeleteOrganization(service: OrganizationService) {
  const deleteOrg = useCallback(
    async (organizationId: string) => {
      await deleteOrganization(service.organizationRepository, organizationId);
    },
    [service]
  );

  return { deleteOrganization: deleteOrg };
}

export function useScenarios(
  service: OrganizationService,
  loadingDispatcher: LoadingDispatcher,
  organizationId: string
) {
  const loadScenarios = useCallback(() => {
    return fetchScenarios(service.scenariosRepository, organizationId);
  }, [service, organizationId]);

  const [scenarios, refreshScenarios] = useSimpleLoader<Scenario[]>(
    loadingDispatcher,
    loadScenarios
  );

  return {
    scenarios,
    refreshScenarios,
  };
}
