# typed: true
# frozen_string_literal: true

class OpenApiSDK::Components::TripleNamespaceConflictTest
  extend ::Crystalline::MetadataFields::ClassMethods
end

class OpenApiSDK::Components::TripleNamespaceConflictTest
  def foo_pet
  end

  def foo_pet=(str_)
  end

  def bar_pet
  end

  def bar_pet=(str_)
  end

  def baz_pet
  end

  def baz_pet=(str_)
  end
end
