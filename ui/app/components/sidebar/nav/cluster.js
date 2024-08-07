/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: BUSL-1.1
 */

import Component from '@glimmer/component';
import { inject as service } from '@ember/service';

export default class SidebarNavClusterComponent extends Component {
  @service currentCluster;
  @service flags;
  @service version;
  @service auth;
  @service namespace;

  get cluster() {
    return this.currentCluster.cluster;
  }

  get isRootNamespace() {
    // should only return true if we're in the true root namespace
    return this.namespace.inRootNamespace && !this.cluster?.hasChrootNamespace;
  }

  get showSync() {
    // Only show sync if cluster is not managed
    return this.flags.managedNamespaceRoot === null;
  }

  get syncBadge() {
    if (this.version.isCommunity) return 'Enterprise';
    if (!this.version.hasSecretsSync) return 'Premium';
    return undefined;
  }
}
