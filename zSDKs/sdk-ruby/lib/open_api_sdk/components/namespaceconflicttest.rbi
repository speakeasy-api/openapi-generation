# typed: true
# frozen_string_literal: true

class OpenApiSDK::Components::NamespaceConflictTest
  extend ::Crystalline::MetadataFields::ClassMethods
end

class OpenApiSDK::Components::NamespaceConflictTest
  def foo_pet
  end

  def foo_pet=(str_)
  end

  def bar_pet
  end

  def bar_pet=(str_)
  end
end
