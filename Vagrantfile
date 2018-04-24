# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|

  config.vm.box = "geerlingguy/ubuntu1604"

  # prefer VMware if available
  config.vm.provider "vmware_workstation"
  config.vm.provider "virtualbox"
  
  # configure providers
  config.vm.provider "vmware_workstation" do |provider, override|
    provider.gui = true
    provider.whitelist_verified = true

    override.vm.synced_folder ".", "/vagrant", {
      :owner => "vagrant",
      :group => "vagrant",
      :mount_options => ["nonempty"]
    }
  end

  config.vm.provider "virtualbox" do |provider, override|
    provider.gui = true

    override.vm.synced_folder ".", "/vagrant", {
      :mount_options => ['dmode=775','fmode=764'],
      :owner => "vagrant",
      :group => "vagrant"
    }
  end
  
  config.ssh.forward_agent = true

  # vagrant-hostmanager
  if Vagrant.has_plugin?("vagrant-hostmanager")
    config.hostmanager.enabled = true
    config.hostmanager.manage_host = true
    config.hostmanager.ignore_private_ip = false
    config.hostmanager.include_offline = true
    config.hostmanager.aliases = "cuddly-potato.test"
    config.hostmanager.ip_resolver = proc do |vm|
      result = ""
      vm.communicate.execute("ifconfig eth0") do |type, data|
        result << data if type == :stdout
      end
      (ip = /^\s*inet .*?(\d+\.\d+\.\d+\.\d+)\s+/.match(result)) && ip[1]
    end
  end

end
